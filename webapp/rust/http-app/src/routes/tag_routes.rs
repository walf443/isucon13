use crate::responses::tag_response::TagResponse;
use axum::extract::State;
use isupipe_core::services::manager::ServiceManager;
use isupipe_core::services::tag_service::TagService;
use isupipe_http_core::error::Error;
use isupipe_http_core::state::AppState;

#[derive(Debug, serde::Serialize)]
pub struct TagsResponse {
    tags: Vec<TagResponse>,
}
pub async fn get_tag_handler<S: ServiceManager>(
    State(AppState { service, .. }): State<AppState<S>>,
) -> Result<axum::Json<TagsResponse>, Error> {
    let tag_models = service.tag_service().find_all().await?;

    let tags = tag_models.into_iter().map(|tag| tag.into()).collect();
    Ok(axum::Json(TagsResponse { tags }))
}
