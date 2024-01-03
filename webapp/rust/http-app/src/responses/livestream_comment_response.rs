use crate::responses::livestream_response::LivestreamResponse;
use crate::responses::user_response::UserResponse;
use isupipe_core::models::livestream_comment::LivestreamComment;
use isupipe_core::services::livestream_service::LivestreamService;
use isupipe_core::services::manager::ServiceManager;
use isupipe_core::services::user_service::UserService;
use isupipe_http_core::responses::ResponseResult;

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
}
