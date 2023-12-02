use crate::utils::fill_livestream_response;
use async_session::{CookieStore, SessionStore};
use axum::extract::{Query, State};
use axum::http::StatusCode;
use axum_extra::extract::SignedCookieJar;
use chrono::{DateTime, NaiveDate, TimeZone, Utc};
use isupipe_core::models::livestream::{Livestream, LivestreamModel};
use isupipe_core::models::livestream_tag::LivestreamTagModel;
use isupipe_core::models::reservation_slot::ReservationSlotModel;
use isupipe_http_core::error::Error;
use isupipe_http_core::state::AppState;
use isupipe_http_core::{verify_user_session, DEFAULT_SESSION_ID_KEY, DEFAULT_USER_ID_KEY};

#[derive(Debug, serde::Deserialize)]
pub struct ReserveLivestreamRequest {
    tags: Vec<i64>,
    title: String,
    description: String,
    playlist_url: String,
    thumbnail_url: String,
    start_at: i64,
    end_at: i64,
}

pub async fn reserve_livestream_handler(
    State(AppState { pool, .. }): State<AppState>,
    jar: SignedCookieJar,
    axum::Json(req): axum::Json<ReserveLivestreamRequest>,
) -> Result<(StatusCode, axum::Json<Livestream>), Error> {
    verify_user_session(&jar).await?;

    if req.tags.iter().any(|&tag_id| tag_id > 103) {
        tracing::error!("unexpected tags: {:?}", req);
    }

    let cookie = jar.get(DEFAULT_SESSION_ID_KEY).ok_or(Error::SessionError)?;
    let sess = CookieStore::new()
        .load_session(cookie.value().to_owned())
        .await?
        .ok_or(Error::SessionError)?;
    let user_id = sess.get(DEFAULT_USER_ID_KEY).ok_or(Error::SessionError)?;

    let mut tx = pool.begin().await?;

    // 2023/11/25 10:00からの１年間の期間内であるかチェック
    let term_start_at = Utc.from_utc_datetime(
        &NaiveDate::from_ymd_opt(2023, 11, 25)
            .unwrap()
            .and_hms_opt(1, 0, 0)
            .unwrap(),
    );
    let term_end_at = Utc.from_utc_datetime(
        &NaiveDate::from_ymd_opt(2024, 11, 25)
            .unwrap()
            .and_hms_opt(1, 0, 0)
            .unwrap(),
    );
    let reserve_start_at = DateTime::from_timestamp(req.start_at, 0).unwrap();
    let reserve_end_at = DateTime::from_timestamp(req.end_at, 0).unwrap();
    if reserve_start_at >= term_end_at || reserve_end_at <= term_start_at {
        return Err(Error::BadRequest("bad reservation time range".into()));
    }

    // 予約枠をみて、予約が可能か調べる
    // NOTE: 並列な予約のoverbooking防止にFOR UPDATEが必要
    let slots: Vec<ReservationSlotModel> = sqlx::query_as(
        "SELECT * FROM reservation_slots WHERE start_at >= ? AND end_at <= ? FOR UPDATE",
    )
    .bind(req.start_at)
    .bind(req.end_at)
    .fetch_all(&mut *tx)
    .await
    .map_err(|e| {
        tracing::warn!("予約枠一覧取得でエラー発生: {e:?}");
        e
    })?;
    for slot in slots {
        let count: i64 = sqlx::query_scalar(
            "SELECT slot FROM reservation_slots WHERE start_at = ? AND end_at = ?",
        )
        .bind(slot.start_at)
        .bind(slot.end_at)
        .fetch_one(&mut *tx)
        .await?;
        tracing::info!(
            "{} ~ {}予約枠の残数 = {}",
            slot.start_at,
            slot.end_at,
            slot.slot
        );
        if count < 1 {
            return Err(Error::BadRequest(
                format!(
                    "予約期間 {} ~ {}に対して、予約区間 {} ~ {}が予約できません",
                    term_start_at.timestamp(),
                    term_end_at.timestamp(),
                    req.start_at,
                    req.end_at
                )
                .into(),
            ));
        }
    }

    sqlx::query("UPDATE reservation_slots SET slot = slot - 1 WHERE start_at >= ? AND end_at <= ?")
        .bind(req.start_at)
        .bind(req.end_at)
        .execute(&mut *tx)
        .await?;

    let rs = sqlx::query("INSERT INTO livestreams (user_id, title, description, playlist_url, thumbnail_url, start_at, end_at) VALUES(?, ?, ?, ?, ?, ?, ?)")
        .bind(user_id)
        .bind(&req.title)
        .bind(&req.description)
        .bind(&req.playlist_url)
        .bind(&req.thumbnail_url)
        .bind(req.start_at)
        .bind(req.end_at)
        .execute(&mut *tx)
        .await?;
    let livestream_id = rs.last_insert_id() as i64;

    // タグ追加
    for tag_id in req.tags {
        sqlx::query("INSERT INTO livestream_tags (livestream_id, tag_id) VALUES (?, ?)")
            .bind(livestream_id)
            .bind(tag_id)
            .execute(&mut *tx)
            .await?;
    }

    let livestream = fill_livestream_response(
        &mut tx,
        LivestreamModel {
            id: livestream_id,
            user_id,
            title: req.title,
            description: req.description,
            playlist_url: req.playlist_url,
            thumbnail_url: req.thumbnail_url,
            start_at: req.start_at,
            end_at: req.end_at,
        },
    )
    .await?;

    tx.commit().await?;

    Ok((StatusCode::CREATED, axum::Json(livestream)))
}

#[derive(Debug, serde::Deserialize)]
pub struct SearchLivestreamsQuery {
    #[serde(default)]
    pub tag: String,
    #[serde(default)]
    pub limit: String,
}

pub async fn search_livestreams_handler(
    State(AppState { pool, .. }): State<AppState>,
    Query(SearchLivestreamsQuery {
              tag: key_tag_name,
              limit,
          }): Query<SearchLivestreamsQuery>,
) -> Result<axum::Json<Vec<Livestream>>, Error> {
    let mut tx = pool.begin().await?;

    let livestream_models: Vec<LivestreamModel> = if key_tag_name.is_empty() {
        // 検索条件なし
        let mut query = "SELECT * FROM livestreams ORDER BY id DESC".to_owned();
        if !limit.is_empty() {
            let limit: i64 = limit
                .parse()
                .map_err(|_| Error::BadRequest("failed to parse limit".into()))?;
            query = format!("{} LIMIT {}", query, limit);
        }
        sqlx::query_as(&query).fetch_all(&mut *tx).await?
    } else {
        // タグによる取得
        let tag_id_list: Vec<i64> = sqlx::query_scalar("SELECT id FROM tags WHERE name = ?")
            .bind(key_tag_name)
            .fetch_all(&mut *tx)
            .await?;

        let mut query_builder = sqlx::query_builder::QueryBuilder::new(
            "SELECT * FROM livestream_tags WHERE tag_id IN (",
        );
        let mut separated = query_builder.separated(", ");
        for tag_id in tag_id_list {
            separated.push_bind(tag_id);
        }
        separated.push_unseparated(") ORDER BY livestream_id DESC");
        let key_tagged_livestreams: Vec<LivestreamTagModel> =
            query_builder.build_query_as().fetch_all(&mut *tx).await?;

        let mut livestream_models = Vec::new();
        for key_tagged_livestream in key_tagged_livestreams {
            let ls = sqlx::query_as("SELECT * FROM livestreams WHERE id = ?")
                .bind(key_tagged_livestream.livestream_id)
                .fetch_one(&mut *tx)
                .await?;
            livestream_models.push(ls);
        }
        livestream_models
    };

    let mut livestreams = Vec::with_capacity(livestream_models.len());
    for livestream_model in livestream_models {
        let livestream = fill_livestream_response(&mut tx, livestream_model).await?;
        livestreams.push(livestream);
    }

    tx.commit().await?;

    Ok(axum::Json(livestreams))
}
