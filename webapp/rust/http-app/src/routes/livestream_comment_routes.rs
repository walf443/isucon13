use crate::utils::fill_livecomment_response;
use async_session::{CookieStore, SessionStore};
use axum::extract::{Path, Query, State};
use axum::http::StatusCode;
use axum_extra::extract::SignedCookieJar;
use chrono::Utc;
use isupipe_core::models::livestream_comment::{LivestreamComment, LivestreamCommentModel};
use isupipe_core::repos::livestream_repository::LivestreamRepository;
use isupipe_core::repos::ng_word_repository::NgWordRepository;
use isupipe_http_core::error::Error;
use isupipe_http_core::state::AppState;
use isupipe_http_core::{verify_user_session, DEFAULT_SESSION_ID_KEY, DEFAULT_USER_ID_KEY};
use isupipe_infra::repos::livestream_repository::LivestreamRepositoryInfra;
use isupipe_infra::repos::ng_word_repository::NgWordRepositoryInfra;

#[derive(Debug, serde::Deserialize)]
pub struct GetLivecommentsQuery {
    #[serde(default)]
    limit: String,
}

pub async fn get_livecomments_handler(
    State(AppState { pool, .. }): State<AppState>,
    jar: SignedCookieJar,
    Path((livestream_id,)): Path<(i64,)>,
    Query(GetLivecommentsQuery { limit }): Query<GetLivecommentsQuery>,
) -> Result<axum::Json<Vec<LivestreamComment>>, Error> {
    verify_user_session(&jar).await?;

    let mut tx = pool.begin().await?;

    let mut query =
        "SELECT * FROM livecomments WHERE livestream_id = ? ORDER BY created_at DESC".to_owned();
    if !limit.is_empty() {
        let limit: i64 = limit.parse().map_err(|_| Error::BadRequest("".into()))?;
        query = format!("{} LIMIT {}", query, limit);
    }

    let livecomment_models: Vec<LivestreamCommentModel> = sqlx::query_as(&query)
        .bind(livestream_id)
        .fetch_all(&mut *tx)
        .await?;

    let mut livecomments = Vec::with_capacity(livecomment_models.len());
    for livecomment_model in livecomment_models {
        let livecomment = fill_livecomment_response(&mut tx, livecomment_model).await?;
        livecomments.push(livecomment);
    }

    tx.commit().await?;

    Ok(axum::Json(livecomments))
}

#[derive(Debug, serde::Deserialize)]
pub struct PostLivecommentRequest {
    pub comment: String,
    pub tip: i64,
}

pub async fn post_livecomment_handler(
    State(AppState { pool, .. }): State<AppState>,
    jar: SignedCookieJar,
    Path((livestream_id,)): Path<(i64,)>,
    axum::Json(req): axum::Json<PostLivecommentRequest>,
) -> Result<(StatusCode, axum::Json<LivestreamComment>), Error> {
    verify_user_session(&jar).await?;

    let cookie = jar.get(DEFAULT_SESSION_ID_KEY).ok_or(Error::SessionError)?;
    let sess = CookieStore::new()
        .load_session(cookie.value().to_owned())
        .await?
        .ok_or(Error::SessionError)?;
    let user_id: i64 = sess.get(DEFAULT_USER_ID_KEY).ok_or(Error::SessionError)?;

    let mut tx = pool.begin().await?;

    let livestream_repo = LivestreamRepositoryInfra {};
    let livestream_model = livestream_repo
        .find(&mut tx, livestream_id)
        .await?
        .ok_or(Error::NotFound("livestream not found".into()))?;

    let ng_word_repo = NgWordRepositoryInfra {};
    // スパム判定
    let ng_words = ng_word_repo
        .find_all_by_livestream_id_and_user_id(
            &mut *tx,
            livestream_model.id,
            livestream_model.user_id,
        )
        .await?;

    for ngword in &ng_words {
        let query = r#"
        SELECT COUNT(*)
        FROM
        (SELECT ? AS text) AS texts
        INNER JOIN
        (SELECT CONCAT('%', ?, '%')	AS pattern) AS patterns
        ON texts.text LIKE patterns.pattern;
        "#;
        let hit_spam: i64 = sqlx::query_scalar(query)
            .bind(&req.comment)
            .bind(&ngword.word)
            .fetch_one(&mut *tx)
            .await?;
        tracing::info!("[hit_spam={}] comment = {}", hit_spam, req.comment);
        if hit_spam >= 1 {
            return Err(Error::BadRequest(
                "このコメントがスパム判定されました".into(),
            ));
        }
    }

    let now = Utc::now().timestamp();

    let rs = sqlx::query(
        "INSERT INTO livecomments (user_id, livestream_id, comment, tip, created_at) VALUES (?, ?, ?, ?, ?)",
    )
        .bind(user_id)
        .bind(livestream_id)
        .bind(&req.comment)
        .bind(req.tip)
        .bind(now)
        .execute(&mut *tx)
        .await?;
    let livecomment_id = rs.last_insert_id() as i64;

    let livecomment = fill_livecomment_response(
        &mut tx,
        LivestreamCommentModel {
            id: livecomment_id,
            user_id,
            livestream_id,
            comment: req.comment,
            tip: req.tip,
            created_at: now,
        },
    )
    .await?;

    tx.commit().await?;

    Ok((StatusCode::CREATED, axum::Json(livecomment)))
}
