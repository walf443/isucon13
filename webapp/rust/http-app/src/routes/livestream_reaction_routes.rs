use crate::responses::reaction_response::ReactionResponse;
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
use isupipe_http_core::error::Error;
use isupipe_http_core::state::AppState;
use isupipe_http_core::{verify_user_session, DEFAULT_SESSION_ID_KEY, DEFAULT_USER_ID_KEY};

#[derive(Debug, serde::Deserialize)]
pub struct GetReactionsQuery {
    #[serde(default)]
    pub limit: String,
}

pub async fn get_reactions_handler<S: ServiceManager>(
    State(AppState { service, pool, .. }): State<AppState<S>>,
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

    let mut tx = pool.begin().await?;
    let mut reactions = Vec::with_capacity(reaction_models.len());
    for reaction_model in reaction_models {
        let reaction = ReactionResponse::build(&mut tx, reaction_model).await?;
        reactions.push(reaction);
    }

    tx.commit().await?;

    Ok(axum::Json(reactions))
}

#[derive(Debug, serde::Deserialize)]
pub struct PostReactionRequest {
    pub emoji_name: String,
}

pub async fn post_reaction_handler<S: ServiceManager>(
    State(AppState { service, pool, .. }): State<AppState<S>>,
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

    let mut tx = pool.begin().await?;

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

    let reaction = ReactionResponse::build(
        &mut tx,
        Reaction {
            id: reaction_id,
            user_id: user_id.clone(),
            livestream_id: livestream_id.clone(),
            emoji_name: req.emoji_name,
            created_at,
        },
    )
    .await?;

    tx.commit().await?;

    Ok((StatusCode::CREATED, axum::Json(reaction)))
}
