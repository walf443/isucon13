use crate::utils::fill_livecomment_response;
use axum::extract::{Path, Query, State};
use axum_extra::extract::SignedCookieJar;
use isupipe_core::models::livestream_comment::{Livecomment, LivecommentModel};
use isupipe_http_core::error::Error;
use isupipe_http_core::state::AppState;
use isupipe_http_core::verify_user_session;

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
) -> Result<axum::Json<Vec<Livecomment>>, Error> {
    verify_user_session(&jar).await?;

    let mut tx = pool.begin().await?;

    let mut query =
        "SELECT * FROM livecomments WHERE livestream_id = ? ORDER BY created_at DESC".to_owned();
    if !limit.is_empty() {
        let limit: i64 = limit.parse().map_err(|_| Error::BadRequest("".into()))?;
        query = format!("{} LIMIT {}", query, limit);
    }

    let livecomment_models: Vec<LivecommentModel> = sqlx::query_as(&query)
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
