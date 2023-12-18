use crate::responses::tag_response::TagResponse;
use axum::extract::State;
use isupipe_core::repos::tag_repository::TagRepository;
use isupipe_http_core::error::Error;
use isupipe_http_core::state::AppState;
use isupipe_infra::repos::tag_repository::TagRepositoryInfra;

#[derive(Debug, serde::Serialize)]
pub struct TagsResponse {
    tags: Vec<TagResponse>,
}
pub async fn get_tag_handler(
    State(AppState { pool, .. }): State<AppState>,
) -> Result<axum::Json<TagsResponse>, Error> {
    let mut tx = pool.begin().await?;

    let tag_repos = TagRepositoryInfra {};
    let tag_models = tag_repos.find_all(&mut tx).await?;

    tx.commit().await?;

    let tags = tag_models.into_iter().map(|tag| tag.into()).collect();
    Ok(axum::Json(TagsResponse { tags }))
}
