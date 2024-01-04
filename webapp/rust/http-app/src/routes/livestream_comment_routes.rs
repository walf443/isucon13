use async_session::{CookieStore, SessionStore};
use axum::extract::{Path, Query, State};
use axum::http::StatusCode;
use axum_extra::extract::SignedCookieJar;
use chrono::Utc;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::livestream_comment::CreateLivestreamComment;
use isupipe_core::models::user::UserId;
use isupipe_core::repos::ng_word_repository::NgWordRepository;
use isupipe_core::services::livestream_comment_service::LivestreamCommentService;
use isupipe_core::services::livestream_service::LivestreamService;
use isupipe_core::services::manager::ServiceManager;
use isupipe_http_core::error::Error;
use isupipe_http_core::responses::livestream_comment_response::LivestreamCommentResponse;
use isupipe_http_core::state::AppState;
use isupipe_http_core::{verify_user_session, DEFAULT_SESSION_ID_KEY, DEFAULT_USER_ID_KEY};
use isupipe_infra::repos::ng_word_repository::NgWordRepositoryInfra;

#[derive(Debug, serde::Deserialize)]
pub struct GetLivestreamCommentsQuery {
    #[serde(default)]
    limit: String,
}

pub async fn get_livestream_comments_handler<S: ServiceManager>(
    State(AppState { service, .. }): State<AppState<S>>,
    jar: SignedCookieJar,
    Path((livestream_id,)): Path<(i64,)>,
    Query(GetLivestreamCommentsQuery { limit }): Query<GetLivestreamCommentsQuery>,
) -> Result<axum::Json<Vec<LivestreamCommentResponse>>, Error> {
    verify_user_session(&jar).await?;

    let livestream_id = LivestreamId::new(livestream_id);

    let limit = if limit.is_empty() {
        None
    } else {
        let limit: i64 = limit.parse().map_err(|_| Error::BadRequest("".into()))?;
        Some(limit)
    };
    let livecomment_models = service
        .livestream_comment_service()
        .find_all_by_livestream_id(&livestream_id, limit)
        .await?;

    let comments =
        LivestreamCommentResponse::bulk_build_by_service(&service, &livecomment_models).await?;

    Ok(axum::Json(comments))
}

#[derive(Debug, serde::Deserialize)]
pub struct PostLivecommentRequest {
    pub comment: String,
    pub tip: i64,
}

pub async fn post_livecomment_handler<S: ServiceManager>(
    State(AppState { service, pool, .. }): State<AppState<S>>,
    jar: SignedCookieJar,
    Path((livestream_id,)): Path<(i64,)>,
    axum::Json(req): axum::Json<PostLivecommentRequest>,
) -> Result<(StatusCode, axum::Json<LivestreamCommentResponse>), Error> {
    verify_user_session(&jar).await?;

    let cookie = jar.get(DEFAULT_SESSION_ID_KEY).ok_or(Error::SessionError)?;
    let sess = CookieStore::new()
        .load_session(cookie.value().to_owned())
        .await?
        .ok_or(Error::SessionError)?;
    let user_id: i64 = sess.get(DEFAULT_USER_ID_KEY).ok_or(Error::SessionError)?;
    let user_id = UserId::new(user_id);
    let livestream_id = LivestreamId::new(livestream_id);

    let mut tx = pool.begin().await?;

    let livestream_model = service
        .livestream_service()
        .find(&livestream_id)
        .await?
        .ok_or(Error::NotFound("livestream not found".into()))?;

    let ng_word_repo = NgWordRepositoryInfra {};
    // スパム判定
    let ng_words = ng_word_repo
        .find_all_by_livestream_id_and_user_id(
            &mut tx,
            &livestream_model.id,
            &livestream_model.user_id,
        )
        .await?;

    for ngword in &ng_words {
        let hit_spam = ng_word_repo
            .count_by_ng_word_in_comment(&mut tx, &ngword.word, &req.comment)
            .await?;
        tracing::info!("[hit_spam={}] comment = {}", hit_spam, req.comment);
        if hit_spam >= 1 {
            return Err(Error::BadRequest(
                "このコメントがスパム判定されました".into(),
            ));
        }
    }

    let now = Utc::now().timestamp();
    let comment = service
        .livestream_comment_service()
        .create(&CreateLivestreamComment {
            user_id: user_id.clone(),
            livestream_id: livestream_id.clone(),
            comment: req.comment.clone(),
            tip: req.tip,
            created_at: now,
        })
        .await?;

    let livecomment = LivestreamCommentResponse::build_by_service(&service, &comment).await?;

    tx.commit().await?;

    Ok((StatusCode::CREATED, axum::Json(livecomment)))
}
