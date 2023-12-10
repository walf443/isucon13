use crate::utils::fill_reaction_response;
use async_session::{CookieStore, SessionStore};
use axum::extract::{Path, Query, State};
use axum::http::StatusCode;
use axum_extra::extract::SignedCookieJar;
use chrono::Utc;
use isupipe_core::models::reaction::{Reaction, ReactionModel};
use isupipe_core::repos::reaction_repository::ReactionRepository;
use isupipe_http_core::error::Error;
use isupipe_http_core::state::AppState;
use isupipe_http_core::{verify_user_session, DEFAULT_SESSION_ID_KEY, DEFAULT_USER_ID_KEY};
use isupipe_infra::repos::reaction_repository::ReactionRepositoryInfra;

#[derive(Debug, serde::Deserialize)]
pub struct GetReactionsQuery {
    #[serde(default)]
    pub limit: String,
}

pub async fn get_reactions_handler(
    State(AppState { pool, .. }): State<AppState>,
    jar: SignedCookieJar,
    Path((livestream_id,)): Path<(i64,)>,
    Query(GetReactionsQuery { limit }): Query<GetReactionsQuery>,
) -> Result<axum::Json<Vec<Reaction>>, Error> {
    verify_user_session(&jar).await?;

    let mut tx = pool.begin().await?;

    let reaction_repo = ReactionRepositoryInfra {};
    let reaction_models = if limit.is_empty() {
        reaction_repo
            .find_all_by_livestream_id(&mut *tx, livestream_id)
            .await?
    } else {
        let limit: i64 = limit.parse().map_err(|_| Error::BadRequest("".into()))?;
        reaction_repo
            .find_all_by_livestream_id_limit(&mut *tx, livestream_id, limit)
            .await?
    };

    let mut reactions = Vec::with_capacity(reaction_models.len());
    for reaction_model in reaction_models {
        let reaction = fill_reaction_response(&mut tx, reaction_model).await?;
        reactions.push(reaction);
    }

    tx.commit().await?;

    Ok(axum::Json(reactions))
}

#[derive(Debug, serde::Deserialize)]
pub struct PostReactionRequest {
    pub emoji_name: String,
}

pub async fn post_reaction_handler(
    State(AppState { pool, .. }): State<AppState>,
    jar: SignedCookieJar,
    Path((livestream_id,)): Path<(i64,)>,
    axum::Json(req): axum::Json<PostReactionRequest>,
) -> Result<(StatusCode, axum::Json<Reaction>), Error> {
    verify_user_session(&jar).await?;

    let cookie = jar.get(DEFAULT_SESSION_ID_KEY).ok_or(Error::SessionError)?;
    let sess = CookieStore::new()
        .load_session(cookie.value().to_owned())
        .await?
        .ok_or(Error::SessionError)?;
    let user_id: i64 = sess.get(DEFAULT_USER_ID_KEY).ok_or(Error::SessionError)?;

    let mut tx = pool.begin().await?;

    let reaction_repo = ReactionRepositoryInfra {};
    let created_at = Utc::now().timestamp();
    let reaction_id = reaction_repo
        .insert(
            &mut *tx,
            user_id,
            livestream_id,
            &req.emoji_name,
            created_at,
        )
        .await?;

    let reaction = fill_reaction_response(
        &mut tx,
        ReactionModel {
            id: reaction_id,
            user_id,
            livestream_id,
            emoji_name: req.emoji_name,
            created_at,
        },
    )
    .await?;

    tx.commit().await?;

    Ok((StatusCode::CREATED, axum::Json(reaction)))
}
