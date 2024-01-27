use crate::responses::tag_response::TagResponse;
use crate::responses::user_response::UserResponse;
use crate::responses::ResponseResult;
use isupipe_core::models::livestream::{Livestream, LivestreamId};
use isupipe_core::services::livestream_tag_service::LivestreamTagService;
use isupipe_core::services::manager::ServiceManager;
use isupipe_core::services::tag_service::TagService;
use isupipe_core::services::user_service::UserService;

#[derive(Debug, serde::Serialize)]
pub struct LivestreamResponse {
    pub id: LivestreamId,
    pub owner: UserResponse,
    pub title: String,
    pub description: String,
    pub playlist_url: String,
    pub thumbnail_url: String,
    pub tags: Vec<TagResponse>,
    pub start_at: i64,
    pub end_at: i64,
}

impl LivestreamResponse {
    pub async fn bulk_build_by_service<S: ServiceManager>(
        service: &S,
        livestream_models: &[Livestream],
    ) -> ResponseResult<Vec<Self>> {
        let mut result = Vec::with_capacity(livestream_models.len());

        for livestream in livestream_models {
            let res = Self::build_by_service(service, livestream).await?;
            result.push(res);
        }

        Ok(result)
    }

    pub async fn build_by_service<S: ServiceManager>(
        service: &S,
        livestream_model: &Livestream,
    ) -> ResponseResult<Self> {
        let owner_model = service
            .user_service()
            .find(&livestream_model.user_id)
            .await?
            .unwrap();
        let owner = UserResponse::build_by_service(service, &owner_model).await?;

        let livestream_tag_models = service
            .livestream_tag_service()
            .find_all_by_livestream_id(&livestream_model.id)
            .await?;

        let mut tags = Vec::with_capacity(livestream_tag_models.len());
        let tag_service = service.tag_service();
        for livestream_tag_model in livestream_tag_models {
            let tag_model = tag_service.find(&livestream_tag_model.tag_id).await?;
            tags.push(TagResponse {
                id: tag_model.id.clone(),
                name: tag_model.name.clone(),
            });
        }

        Ok(Self {
            id: livestream_model.id.clone(),
            owner,
            title: livestream_model.title.clone(),
            tags,
            description: livestream_model.description.clone(),
            playlist_url: livestream_model.playlist_url.clone(),
            thumbnail_url: livestream_model.thumbnail_url.clone(),
            start_at: livestream_model.start_at,
            end_at: livestream_model.end_at,
        })
    }
}
