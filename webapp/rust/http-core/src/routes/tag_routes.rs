use crate::error::Error;
use crate::responses::tag_response::TagResponse;
use crate::state::AppState;
use axum::extract::State;
use isupipe_core::services::manager::ServiceManager;
use isupipe_core::services::tag_service::TagService;

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
