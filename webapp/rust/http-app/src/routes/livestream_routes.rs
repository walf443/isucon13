use crate::utils::fill_livestream_response;
use async_session::{CookieStore, SessionStore};
use axum::extract::{Path, Query, State};
use axum::http::StatusCode;
use axum_extra::extract::SignedCookieJar;
use chrono::{DateTime, NaiveDate, TimeZone, Utc};
use isupipe_core::models::livestream::{Livestream, LivestreamModel};
use isupipe_core::models::livestream_comment::LivestreamCommentModel;
use isupipe_core::models::livestream_ranking_entry::LivestreamRankingEntry;
use isupipe_core::models::livestream_statistics::LivestreamStatistics;
use isupipe_core::models::livestream_tag::LivestreamTagModel;
use isupipe_core::models::mysql_decimal::MysqlDecimal;
use isupipe_core::models::ng_word::NgWord;
use isupipe_core::models::reservation_slot::ReservationSlotModel;
use isupipe_core::repos::livestream_comment_report_repository::LivestreamCommentReportRepository;
use isupipe_core::repos::livestream_viewers_history_repository::LivestreamViewersHistoryRepository;
use isupipe_core::repos::ng_word_repository::NgWordRepository;
use isupipe_core::repos::reaction_repository::ReactionRepository;
use isupipe_core::repos::tag_repository::TagRepository;
use isupipe_http_core::error::Error;
use isupipe_http_core::state::AppState;
use isupipe_http_core::{verify_user_session, DEFAULT_SESSION_ID_KEY, DEFAULT_USER_ID_KEY};
use isupipe_infra::repos::livestream_comment_report_repository::LivestreamCommentReportRepositoryInfra;
use isupipe_infra::repos::livestream_viewers_history_repository::LivestreamViewersHistoryRepositoryInfra;
use isupipe_infra::repos::ng_word_repository::NgWordRepositoryInfra;
use isupipe_infra::repos::reaction_repository::ReactionRepositoryInfra;
use isupipe_infra::repos::tag_repository::TagRepositoryInfra;

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
        let tag_repo = TagRepositoryInfra {};
        let tag_id_list = tag_repo.find_ids_by_name(&mut *tx, &key_tag_name).await?;

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
pub async fn get_my_livestreams_handler(
    State(AppState { pool, .. }): State<AppState>,
    jar: SignedCookieJar,
) -> Result<axum::Json<Vec<Livestream>>, Error> {
    verify_user_session(&jar).await?;

    let cookie = jar.get(DEFAULT_SESSION_ID_KEY).ok_or(Error::SessionError)?;
    let sess = CookieStore::new()
        .load_session(cookie.value().to_owned())
        .await?
        .ok_or(Error::SessionError)?;
    let user_id: i64 = sess.get(DEFAULT_USER_ID_KEY).ok_or(Error::SessionError)?;

    let mut tx = pool.begin().await?;

    let livestream_models: Vec<LivestreamModel> =
        sqlx::query_as("SELECT * FROM livestreams WHERE user_id = ?")
            .bind(user_id)
            .fetch_all(&mut *tx)
            .await?;
    let mut livestreams = Vec::with_capacity(livestream_models.len());
    for livestream_model in livestream_models {
        let livestream = fill_livestream_response(&mut tx, livestream_model).await?;
        livestreams.push(livestream);
    }

    tx.commit().await?;

    Ok(axum::Json(livestreams))
}

pub async fn get_livestream_handler(
    State(AppState { pool, .. }): State<AppState>,
    jar: SignedCookieJar,
    Path((livestream_id,)): Path<(i64,)>,
) -> Result<axum::Json<Livestream>, Error> {
    verify_user_session(&jar).await?;

    let mut tx = pool.begin().await?;

    let livestream_model: LivestreamModel =
        sqlx::query_as("SELECT * FROM livestreams WHERE id = ?")
            .bind(livestream_id)
            .fetch_optional(&mut *tx)
            .await?
            .ok_or(Error::NotFound(
                "not found livestream that has the given id".into(),
            ))?;

    let livestream = fill_livestream_response(&mut tx, livestream_model).await?;

    tx.commit().await?;

    Ok(axum::Json(livestream))
}
pub async fn get_ngwords(
    State(AppState { pool, .. }): State<AppState>,
    jar: SignedCookieJar,
    Path((livestream_id,)): Path<(i64,)>,
) -> Result<axum::Json<Vec<NgWord>>, Error> {
    verify_user_session(&jar).await?;

    let cookie = jar.get(DEFAULT_SESSION_ID_KEY).ok_or(Error::SessionError)?;
    let sess = CookieStore::new()
        .load_session(cookie.value().to_owned())
        .await?
        .ok_or(Error::SessionError)?;
    let user_id: i64 = sess.get(DEFAULT_USER_ID_KEY).ok_or(Error::SessionError)?;

    let mut tx = pool.begin().await?;

    let ng_word_repo = NgWordRepositoryInfra {};
    let ng_words = ng_word_repo
        .find_all_by_livestream_id_and_user_id_order_by_created_at(&mut *tx, livestream_id, user_id)
        .await?;

    tx.commit().await?;

    Ok(axum::Json(ng_words))
}
#[derive(Debug, serde::Deserialize)]
pub struct ModerateRequest {
    ng_word: String,
}

#[derive(Debug, serde::Serialize)]
pub struct ModerateResponse {
    word_id: i64,
}

// NGワードを登録
pub async fn moderate_handler(
    State(AppState { pool, .. }): State<AppState>,
    jar: SignedCookieJar,
    Path((livestream_id,)): Path<(i64,)>,
    axum::Json(req): axum::Json<ModerateRequest>,
) -> Result<(StatusCode, axum::Json<ModerateResponse>), Error> {
    verify_user_session(&jar).await?;

    let cookie = jar.get(DEFAULT_SESSION_ID_KEY).ok_or(Error::SessionError)?;
    let sess = CookieStore::new()
        .load_session(cookie.value().to_owned())
        .await?
        .ok_or(Error::SessionError)?;
    let user_id: i64 = sess.get(DEFAULT_USER_ID_KEY).ok_or(Error::SessionError)?;

    let mut tx = pool.begin().await?;

    // 配信者自身の配信に対するmoderateなのかを検証
    let owned_livestreams: Vec<LivestreamModel> =
        sqlx::query_as("SELECT * FROM livestreams WHERE id = ? AND user_id = ?")
            .bind(livestream_id)
            .bind(user_id)
            .fetch_all(&mut *tx)
            .await?;
    if owned_livestreams.is_empty() {
        return Err(Error::BadRequest(
            "A streamer can't moderate livestreams that other streamers own".into(),
        ));
    }

    let created_at = Utc::now().timestamp();
    let ng_word_repo = NgWordRepositoryInfra {};
    let word_id = ng_word_repo
        .insert(&mut *tx, user_id, livestream_id, &req.ng_word, created_at)
        .await?;

    let ng_words = ng_word_repo
        .find_all_by_livestream_id(&mut *tx, livestream_id)
        .await?;

    // NGワードにヒットする過去の投稿も全削除する
    for ngword in ng_words {
        // ライブコメント一覧取得
        let livecomments: Vec<LivestreamCommentModel> =
            sqlx::query_as("SELECT * FROM livecomments")
                .fetch_all(&mut *tx)
                .await?;

        for livecomment in livecomments {
            let query = r#"
            DELETE FROM livecomments
            WHERE
            id = ? AND
            livestream_id = ? AND
            (SELECT COUNT(*)
            FROM
            (SELECT ? AS text) AS texts
            INNER JOIN
            (SELECT CONCAT('%', ?, '%')	AS pattern) AS patterns
            ON texts.text LIKE patterns.pattern) >= 1
            "#;
            sqlx::query(query)
                .bind(livecomment.id)
                .bind(livestream_id)
                .bind(livecomment.comment)
                .bind(&ngword.word)
                .execute(&mut *tx)
                .await?;
        }
    }

    tx.commit().await?;

    Ok((
        StatusCode::CREATED,
        axum::Json(ModerateResponse { word_id }),
    ))
}

// viewerテーブルの廃止
pub async fn enter_livestream_handler(
    State(AppState { pool, .. }): State<AppState>,
    jar: SignedCookieJar,
    Path((livestream_id,)): Path<(i64,)>,
) -> Result<(), Error> {
    verify_user_session(&jar).await?;

    let cookie = jar.get(DEFAULT_SESSION_ID_KEY).ok_or(Error::SessionError)?;
    let sess = CookieStore::new()
        .load_session(cookie.value().to_owned())
        .await?
        .ok_or(Error::SessionError)?;
    let user_id: i64 = sess.get(DEFAULT_USER_ID_KEY).ok_or(Error::SessionError)?;

    let created_at = Utc::now().timestamp();
    let mut conn = pool.acquire().await?;
    let history_repo = LivestreamViewersHistoryRepositoryInfra {};
    history_repo
        .insert(&mut conn, livestream_id, user_id, created_at)
        .await?;

    Ok(())
}
pub async fn exit_livestream_handler(
    State(AppState { pool, .. }): State<AppState>,
    jar: SignedCookieJar,
    Path((livestream_id,)): Path<(i64,)>,
) -> Result<(), Error> {
    verify_user_session(&jar).await?;

    let cookie = jar.get(DEFAULT_SESSION_ID_KEY).ok_or(Error::SessionError)?;
    let sess = CookieStore::new()
        .load_session(cookie.value().to_owned())
        .await?
        .ok_or(Error::SessionError)?;
    let user_id: i64 = sess.get(DEFAULT_USER_ID_KEY).ok_or(Error::SessionError)?;

    let history_repo = LivestreamViewersHistoryRepositoryInfra {};
    let mut conn = pool.acquire().await?;
    history_repo
        .delete_by_livestream_id_and_user_id(&mut conn, livestream_id, user_id)
        .await?;

    Ok(())
}
pub async fn get_livestream_statistics_handler(
    State(AppState { pool, .. }): State<AppState>,
    jar: SignedCookieJar,
    Path((livestream_id,)): Path<(i64,)>,
) -> Result<axum::Json<LivestreamStatistics>, Error> {
    verify_user_session(&jar).await?;

    let mut tx = pool.begin().await?;

    let _: LivestreamModel = sqlx::query_as("SELECT * FROM livestreams WHERE id = ?")
        .bind(livestream_id)
        .fetch_optional(&mut *tx)
        .await?
        .ok_or(Error::BadRequest("".into()))?;

    let livestreams: Vec<LivestreamModel> = sqlx::query_as("SELECT * FROM livestreams")
        .fetch_all(&mut *tx)
        .await?;

    // ランク算出
    let mut ranking = Vec::new();
    let reaction_repo = ReactionRepositoryInfra {};
    for livestream in livestreams {
        let reactions = reaction_repo
            .count_by_livestream_id(&mut *tx, livestream.id)
            .await?;

        let MysqlDecimal(total_tips) = sqlx::query_scalar("SELECT IFNULL(SUM(l2.tip), 0) FROM livestreams l INNER JOIN livecomments l2 ON l.id = l2.livestream_id WHERE l.id = ?")
            .bind(livestream.id)
            .fetch_one(&mut *tx)
            .await?;

        let score = reactions + total_tips;
        ranking.push(LivestreamRankingEntry {
            livestream_id: livestream.id,
            score,
        })
    }
    ranking.sort_by(|a, b| {
        a.score
            .cmp(&b.score)
            .then_with(|| a.livestream_id.cmp(&b.livestream_id))
    });

    let rpos = ranking
        .iter()
        .rposition(|entry| entry.livestream_id == livestream_id)
        .unwrap();
    let rank = (ranking.len() - rpos) as i64;

    // 視聴者数算出
    let history_repo = LivestreamViewersHistoryRepositoryInfra {};
    let viewers_count = history_repo
        .count_by_livestream_id(&mut tx, livestream_id)
        .await?;

    // 最大チップ額
    let MysqlDecimal(max_tip) = sqlx::query_scalar("SELECT IFNULL(MAX(tip), 0) FROM livestreams l INNER JOIN livecomments l2 ON l2.livestream_id = l.id WHERE l.id = ?")
        .bind(livestream_id)
        .fetch_one(&mut *tx)
        .await?;

    // リアクション数
    let reaction_repo = ReactionRepositoryInfra {};
    let total_reactions = reaction_repo
        .count_by_livestream_id(&mut *tx, livestream_id)
        .await?;

    // スパム報告数
    let report_repo = LivestreamCommentReportRepositoryInfra {};
    let total_reports = report_repo
        .count_by_livestream_id(&mut *tx, livestream_id)
        .await?;

    tx.commit().await?;

    Ok(axum::Json(LivestreamStatistics {
        rank,
        viewers_count,
        max_tip,
        total_reactions,
        total_reports,
    }))
}
