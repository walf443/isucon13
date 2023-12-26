use crate::responses::livestream_response::LivestreamResponse;
use crate::responses::user_response::UserResponse;
use isupipe_core::db::DBConn;
use isupipe_core::models::reaction::Reaction;
use isupipe_core::repos::livestream_repository::LivestreamRepository;
use isupipe_core::repos::user_repository::UserRepository;
use isupipe_http_core::responses::ResponseResult;
use isupipe_infra::repos::livestream_repository::LivestreamRepositoryInfra;
use isupipe_infra::repos::user_repository::UserRepositoryInfra;

#[derive(Debug, serde::Serialize)]
pub struct ReactionResponse {
    pub id: i64,
    pub emoji_name: String,
    pub user: UserResponse,
    pub livestream: LivestreamResponse,
    pub created_at: i64,
}

impl ReactionResponse {
    pub async fn build(conn: &mut DBConn, reaction_model: Reaction) -> ResponseResult<Self> {
        let user_repo = UserRepositoryInfra {};
        let user_model = user_repo
            .find(conn, reaction_model.user_id.get())
            .await?
            .unwrap();
        let user = UserResponse::build(conn, user_model).await?;

        let livestream_repo = LivestreamRepositoryInfra {};
        let livestream_model = livestream_repo
            .find(conn, reaction_model.livestream_id.get())
            .await?
            .unwrap();
        let livestream = LivestreamResponse::build(conn, livestream_model).await?;

        Ok(Self {
            id: reaction_model.id.get(),
            emoji_name: reaction_model.emoji_name,
            user,
            livestream,
            created_at: reaction_model.created_at,
        })
    }
}
