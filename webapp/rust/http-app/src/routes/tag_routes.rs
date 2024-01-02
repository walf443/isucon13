use crate::responses::tag_response::TagResponse;
use axum::extract::State;
use isupipe_core::services::manager::ServiceManager;
use isupipe_core::services::tag_service::{HaveTagService, TagService};
use isupipe_http_core::error::Error;
use isupipe_http_core::state::AppState;
use isupipe_infra::services::manager::ServiceManagerInfra;

#[derive(Debug, serde::Serialize)]
pub struct TagsResponse {
    tags: Vec<TagResponse>,
}
pub async fn get_tag_handler<S: ServiceManager>(
    State(AppState { pool, .. }): State<AppState<S>>,
) -> Result<axum::Json<TagsResponse>, Error> {
    let service = ServiceManagerInfra::new(pool.clone());

    let tag_models = service.tag_service().find_all().await?;

    let tags = tag_models.into_iter().map(|tag| tag.into()).collect();
    Ok(axum::Json(TagsResponse { tags }))
}
