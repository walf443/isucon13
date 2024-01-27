use crate::responses::livestream_response::LivestreamResponse;
use crate::responses::user_response::UserResponse;
use crate::responses::ResponseResult;
use isupipe_core::models::reaction::{Reaction, ReactionId};
use isupipe_core::services::livestream_service::LivestreamService;
use isupipe_core::services::manager::ServiceManager;
use isupipe_core::services::user_service::UserService;

#[derive(Debug, serde::Serialize)]
pub struct ReactionResponse {
    pub id: ReactionId,
    pub emoji_name: String,
    pub user: UserResponse,
    pub livestream: LivestreamResponse,
    pub created_at: i64,
}

impl ReactionResponse {
    pub async fn bulk_build_by_service<S: ServiceManager>(
        service: &S,
        reactions: &[Reaction],
    ) -> ResponseResult<Vec<Self>> {
        let mut result = Vec::with_capacity(reactions.len());

        for reaction in reactions {
            let res = Self::build_by_service(service, reaction).await?;
            result.push(res)
        }

        Ok(result)
    }

    pub async fn build_by_service<S: ServiceManager>(
        service: &S,
        reaction_model: &Reaction,
    ) -> ResponseResult<Self> {
        let user_model = service
            .user_service()
            .find(&reaction_model.user_id)
            .await?
            .unwrap();
        let user = UserResponse::build_by_service(service, &user_model).await?;

        let livestream_model = service
            .livestream_service()
            .find(&reaction_model.livestream_id)
            .await?
            .unwrap();
        let livestream = LivestreamResponse::build_by_service(service, &livestream_model).await?;

        Ok(Self {
            id: reaction_model.id.clone(),
            emoji_name: reaction_model.emoji_name.clone(),
            user,
            livestream,
            created_at: reaction_model.created_at,
        })
    }
}
