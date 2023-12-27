use crate::db::HaveDBPool;
use crate::models::reaction::{CreateReaction, ReactionId};
use crate::repos::reaction_repository::{HaveReactionRepository, ReactionRepository};
use crate::services::ServiceResult;
use async_trait::async_trait;

#[async_trait]
pub trait ReactionService {
    async fn create(&self, reaction: &CreateReaction) -> ServiceResult<ReactionId>;
}

pub trait HaveReactionService {
    type Service: ReactionService;

    fn reaction_service(&self) -> &Self::Service;
}

pub trait ReactionServiceImpl: Sync + HaveDBPool + HaveReactionRepository {}

#[async_trait]
impl<T: ReactionServiceImpl> ReactionService for T {
    async fn create(&self, reaction: &CreateReaction) -> ServiceResult<ReactionId> {
        let mut conn = self.get_db_pool().acquire().await?;
        let reaction_id = self.reaction_repo().create(&mut *conn, reaction).await?;

        Ok(reaction_id)
    }
}
