use crate::db::HaveDBPool;
use crate::models::livestream::LivestreamId;
use crate::models::reaction::{CreateReaction, Reaction, ReactionId};
use crate::repos::reaction_repository::{HaveReactionRepository, ReactionRepository};
use crate::services::ServiceResult;
use async_trait::async_trait;

#[async_trait]
pub trait ReactionService {
    async fn create(&self, reaction: &CreateReaction) -> ServiceResult<ReactionId>;

    async fn find_all_by_livestream_id_limit(
        &self,
        livestream_id: &LivestreamId,
        limit: Option<i64>,
    ) -> ServiceResult<Vec<Reaction>>;
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

    async fn find_all_by_livestream_id_limit(
        &self,
        livestream_id: &LivestreamId,
        limit: Option<i64>,
    ) -> ServiceResult<Vec<Reaction>> {
        let mut conn = self.get_db_pool().acquire().await?;
        let result = match limit {
            None => {
                self.reaction_repo()
                    .find_all_by_livestream_id(&mut *conn, livestream_id)
                    .await
            }
            Some(limit) => {
                self.reaction_repo()
                    .find_all_by_livestream_id_limit(&mut *conn, livestream_id, limit)
                    .await
            }
        }?;

        Ok(result)
    }
}
