use axum::extract::State;
use isupipe_core::models::tag::{Tag, TagModel};
use isupipe_http_core::error::Error;
use isupipe_http_core::state::AppState;

#[derive(Debug, serde::Serialize)]
pub struct TagsResponse {
    tags: Vec<Tag>,
}
pub async fn get_tag_handler(
    State(AppState { pool, .. }): State<AppState>,
) -> Result<axum::Json<TagsResponse>, Error> {
    let mut tx = pool.begin().await?;

    let tag_models: Vec<TagModel> = sqlx::query_as("SELECT * FROM tags")
        .fetch_all(&mut *tx)
        .await?;

    tx.commit().await?;

    let tags = tag_models
        .into_iter()
        .map(|tag| Tag {
            id: tag.id,
            name: tag.name,
        })
        .collect();
    Ok(axum::Json(TagsResponse { tags }))
}
