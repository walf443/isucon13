use crate::responses::user_response::UserResponse;
use isupipe_core::db::DBConn;
use isupipe_core::models::livestream::Livestream;
use isupipe_core::repos::livestream_tag_repository::LivestreamTagRepository;
use isupipe_core::repos::tag_repository::TagRepository;
use isupipe_core::repos::user_repository::UserRepository;
use isupipe_core::services::livestream_tag_service::LivestreamTagService;
use isupipe_core::services::manager::ServiceManager;
use isupipe_core::services::tag_service::TagService;
use isupipe_core::services::user_service::UserService;
use isupipe_http_core::responses::tag_response::TagResponse;
use isupipe_http_core::responses::ResponseResult;
use isupipe_infra::repos::livestream_tag_repository::LivestreamTagRepositoryInfra;
use isupipe_infra::repos::tag_repository::TagRepositoryInfra;
use isupipe_infra::repos::user_repository::UserRepositoryInfra;

#[derive(Debug, serde::Serialize)]
pub struct LivestreamResponse {
    pub id: i64,
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
                id: tag_model.id.get(),
                name: tag_model.name,
            });
        }

        Ok(Self {
            id: livestream_model.id.get(),
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

    pub async fn build(conn: &mut DBConn, livestream_model: Livestream) -> ResponseResult<Self> {
        let user_repo = UserRepositoryInfra {};
        let owner_model = user_repo
            .find(conn, &livestream_model.user_id)
            .await?
            .unwrap();
        let owner = UserResponse::build(conn, owner_model).await?;

        let livestream_tag_repo = LivestreamTagRepositoryInfra {};
        let livestream_tag_models = livestream_tag_repo
            .find_all_by_livestream_id(conn, &livestream_model.id)
            .await?;

        let tag_repo = TagRepositoryInfra {};
        let mut tags = Vec::with_capacity(livestream_tag_models.len());
        for livestream_tag_model in livestream_tag_models {
            let tag_model = tag_repo.find(conn, &livestream_tag_model.tag_id).await?;
            tags.push(TagResponse {
                id: tag_model.id.get(),
                name: tag_model.name,
            });
        }

        Ok(Self {
            id: livestream_model.id.get(),
            owner,
            title: livestream_model.title,
            tags,
            description: livestream_model.description,
            playlist_url: livestream_model.playlist_url,
            thumbnail_url: livestream_model.thumbnail_url,
            start_at: livestream_model.start_at,
            end_at: livestream_model.end_at,
        })
    }
}
