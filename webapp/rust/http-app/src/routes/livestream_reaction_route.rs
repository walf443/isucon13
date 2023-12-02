use axum::extract::{Path, Query, State};
use axum_extra::extract::SignedCookieJar;
use isupipe_core::models::reaction::{Reaction, ReactionModel};
use isupipe_http_core::error::Error;
use isupipe_http_core::state::AppState;
use isupipe_http_core::verify_user_session;
use crate::utils::fill_reaction_response;

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

    let mut query =
        "SELECT * FROM reactions WHERE livestream_id = ? ORDER BY created_at DESC".to_owned();
    if !limit.is_empty() {
        let limit: i64 = limit.parse().map_err(|_| Error::BadRequest("".into()))?;
        query = format!("{} LIMIT {}", query, limit);
    }

    let reaction_models: Vec<ReactionModel> = sqlx::query_as(&query)
        .bind(livestream_id)
        .fetch_all(&mut *tx)
        .await?;

    let mut reactions = Vec::with_capacity(reaction_models.len());
    for reaction_model in reaction_models {
        let reaction = fill_reaction_response(&mut tx, reaction_model).await?;
        reactions.push(reaction);
    }

    tx.commit().await?;

    Ok(axum::Json(reactions))
}
