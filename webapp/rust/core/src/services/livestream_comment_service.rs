use crate::db::HaveDBPool;
use crate::models::livestream::LivestreamId;
use crate::models::livestream_comment::{LivestreamComment, LivestreamCommentId};
use crate::repos::livestream_comment_repository::{
    HaveLivestreamCommentRepository, LivestreamCommentRepository,
};
use crate::services::ServiceResult;
use async_trait::async_trait;

#[async_trait]
pub trait LivestreamCommentService {
    async fn find(
        &self,
        livestream_comment_id: &LivestreamCommentId,
    ) -> ServiceResult<Option<LivestreamComment>>;

    async fn find_all_by_livestream_id(
        &self,
        livestream_id: &LivestreamId,
        limit: Option<i64>,
    ) -> ServiceResult<Vec<LivestreamComment>>;
    async fn get_sum_tip(&self) -> ServiceResult<i64>;
}

pub trait HaveLivestreamCommentService {
    type Service: LivestreamCommentService;
    fn livestream_comment_service(&self) -> &Self::Service;
}

pub trait LivestreamCommentServiceImpl:
    Sync + HaveDBPool + HaveLivestreamCommentRepository
{
}

#[async_trait]
impl<T: LivestreamCommentServiceImpl> LivestreamCommentService for T {
    async fn find(
        &self,
        livestream_comment_id: &LivestreamCommentId,
    ) -> ServiceResult<Option<LivestreamComment>> {
        let mut conn = self.get_db_pool().acquire().await?;
        let comment = self
            .livestream_comment_repo()
            .find(&mut conn, livestream_comment_id)
            .await?;
        Ok(comment)
    }

    async fn find_all_by_livestream_id(
        &self,
        livestream_id: &LivestreamId,
        limit: Option<i64>,
    ) -> ServiceResult<Vec<LivestreamComment>> {
        let mut conn = self.get_db_pool().acquire().await?;

        let comments = match limit {
            None => {
                self.livestream_comment_repo()
                    .find_all_by_livestream_id_order_by_created_at(&mut conn, livestream_id)
                    .await?
            }
            Some(limit) => {
                self.livestream_comment_repo()
                    .find_all_by_livestream_id_order_by_created_at_limit(
                        &mut conn,
                        livestream_id,
                        limit,
                    )
                    .await?
            }
        };

        Ok(comments)
    }

    async fn get_sum_tip(&self) -> ServiceResult<i64> {
        let mut conn = self.get_db_pool().acquire().await?;
        let sum_tip = self
            .livestream_comment_repo()
            .get_sum_tip(&mut conn)
            .await?;
        Ok(sum_tip)
    }
}
