use crate::error::Error;
use crate::responses::reaction_response::ReactionResponse;
use crate::state::AppState;
use crate::{verify_user_session, DEFAULT_SESSION_ID_KEY, DEFAULT_USER_ID_KEY};
use async_session::{CookieStore, SessionStore};
use axum::extract::{Path, Query, State};
use axum::http::StatusCode;
use axum_extra::extract::SignedCookieJar;
use chrono::Utc;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::reaction::{CreateReaction, Reaction};
use isupipe_core::models::user::UserId;
use isupipe_core::services::manager::ServiceManager;
use isupipe_core::services::reaction_service::ReactionService;

#[derive(Debug, serde::Deserialize)]
pub struct GetReactionsQuery {
    #[serde(default)]
    pub limit: String,
}

pub async fn get_reactions_handler<S: ServiceManager>(
    State(AppState { service, .. }): State<AppState<S>>,
    jar: SignedCookieJar,
    Path((livestream_id,)): Path<(i64,)>,
    Query(GetReactionsQuery { limit }): Query<GetReactionsQuery>,
) -> Result<axum::Json<Vec<ReactionResponse>>, Error> {
    verify_user_session(&jar).await?;
    let livestream_id = LivestreamId::new(livestream_id);

    let limit = if limit.is_empty() {
        None
    } else {
        let limit: i64 = limit.parse().map_err(|_| Error::BadRequest("".into()))?;
        Some(limit)
    };
    let reaction_models = service
        .reaction_service()
        .find_all_by_livestream_id_limit(&livestream_id, limit)
        .await?;

    let reactions = ReactionResponse::bulk_build_by_service(&service, &reaction_models).await?;

    Ok(axum::Json(reactions))
}

#[derive(Debug, serde::Deserialize)]
pub struct PostReactionRequest {
    pub emoji_name: String,
}

pub async fn post_reaction_handler<S: ServiceManager>(
    State(AppState { service, .. }): State<AppState<S>>,
    jar: SignedCookieJar,
    Path((livestream_id,)): Path<(i64,)>,
    axum::Json(req): axum::Json<PostReactionRequest>,
) -> Result<(StatusCode, axum::Json<ReactionResponse>), Error> {
    verify_user_session(&jar).await?;

    let cookie = jar.get(DEFAULT_SESSION_ID_KEY).ok_or(Error::SessionError)?;
    let sess = CookieStore::new()
        .load_session(cookie.value().to_owned())
        .await?
        .ok_or(Error::SessionError)?;
    let user_id: i64 = sess.get(DEFAULT_USER_ID_KEY).ok_or(Error::SessionError)?;
    let user_id = UserId::new(user_id);

    let livestream_id = LivestreamId::new(livestream_id);

    let created_at = Utc::now().timestamp();
    let reaction_id = service
        .reaction_service()
        .create(&CreateReaction {
            emoji_name: req.emoji_name.clone(),
            user_id: user_id.clone(),
            livestream_id: livestream_id.clone(),
            created_at,
        })
        .await?;

    let reaction = ReactionResponse::build_by_service(
        &service,
        &Reaction {
            id: reaction_id,
            user_id: user_id.clone(),
            livestream_id: livestream_id.clone(),
            emoji_name: req.emoji_name,
            created_at,
        },
    )
    .await?;

    Ok((StatusCode::CREATED, axum::Json(reaction)))
}
