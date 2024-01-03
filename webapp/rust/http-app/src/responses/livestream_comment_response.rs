use crate::responses::livestream_response::LivestreamResponse;
use crate::responses::user_response::UserResponse;
use isupipe_core::db::DBConn;
use isupipe_core::models::livestream_comment::LivestreamComment;
use isupipe_core::repos::livestream_repository::LivestreamRepository;
use isupipe_core::repos::user_repository::UserRepository;
use isupipe_core::services::livestream_service::LivestreamService;
use isupipe_core::services::manager::ServiceManager;
use isupipe_core::services::user_service::UserService;
use isupipe_http_core::responses::ResponseResult;
use isupipe_infra::repos::livestream_repository::LivestreamRepositoryInfra;
use isupipe_infra::repos::user_repository::UserRepositoryInfra;

#[derive(Debug, serde::Serialize)]
pub struct LivestreamCommentResponse {
    pub id: i64,
    pub user: UserResponse,
    pub livestream: LivestreamResponse,
    pub comment: String,
    pub tip: i64,
    pub created_at: i64,
}

impl LivestreamCommentResponse {
    pub async fn build_by_service<S: ServiceManager>(
        service: &S,
        livecomment_model: &LivestreamComment,
    ) -> ResponseResult<Self> {
        let comment_owner_model = service
            .user_service()
            .find(&livecomment_model.user_id)
            .await?
            .unwrap();
        let comment_owner = UserResponse::build_by_service(service, &comment_owner_model).await?;

        let livestream_service = service.livestream_service();
        let livestream_model = livestream_service
            .find(&livecomment_model.livestream_id)
            .await?
            .unwrap();
        let livestream = LivestreamResponse::build_by_service(service, &livestream_model).await?;

        Ok(Self {
            id: livecomment_model.id.get(),
            user: comment_owner,
            livestream,
            comment: livecomment_model.comment.clone(),
            tip: livecomment_model.tip,
            created_at: livecomment_model.created_at,
        })
    }
    pub async fn build(
        conn: &mut DBConn,
        livecomment_model: LivestreamComment,
    ) -> ResponseResult<Self> {
        let user_repo = UserRepositoryInfra {};
        let comment_owner_model = user_repo
            .find(conn, &livecomment_model.user_id)
            .await?
            .unwrap();
        let comment_owner = UserResponse::build(conn, comment_owner_model).await?;

        let livestream_repo = LivestreamRepositoryInfra {};
        let livestream_model = livestream_repo
            .find(conn, &livecomment_model.livestream_id)
            .await?
            .unwrap();
        let livestream = LivestreamResponse::build(conn, livestream_model).await?;

        Ok(Self {
            id: livecomment_model.id.get(),
            user: comment_owner,
            livestream,
            comment: livecomment_model.comment,
            tip: livecomment_model.tip,
            created_at: livecomment_model.created_at,
        })
    }
}
