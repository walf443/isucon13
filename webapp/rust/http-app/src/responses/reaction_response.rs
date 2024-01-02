use crate::responses::livestream_response::LivestreamResponse;
use crate::responses::user_response::UserResponse;
use isupipe_core::models::reaction::Reaction;
use isupipe_core::services::livestream_service::LivestreamService;
use isupipe_core::services::manager::ServiceManager;
use isupipe_core::services::user_service::UserService;
use isupipe_http_core::responses::ResponseResult;

#[derive(Debug, serde::Serialize)]
pub struct ReactionResponse {
    pub id: i64,
    pub emoji_name: String,
    pub user: UserResponse,
    pub livestream: LivestreamResponse,
    pub created_at: i64,
}

impl ReactionResponse {
    pub async fn build_by_service<S: ServiceManager>(
        service: &S,
        reaction_model: Reaction,
    ) -> ResponseResult<Self> {
        let user_model = service
            .user_service()
            .find(&reaction_model.user_id)
            .await?
            .unwrap();
        let user = UserResponse::build_by_service(service, user_model).await?;

        let livestream_model = service
            .livestream_service()
            .find(&reaction_model.livestream_id)
            .await?
            .unwrap();
        let livestream = LivestreamResponse::build_by_service(service, livestream_model).await?;

        Ok(Self {
            id: reaction_model.id.get(),
            emoji_name: reaction_model.emoji_name,
            user,
            livestream,
            created_at: reaction_model.created_at,
        })
    }
}
