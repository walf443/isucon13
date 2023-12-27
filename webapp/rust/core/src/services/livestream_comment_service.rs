use crate::db::HaveDBPool;
use crate::repos::livestream_comment_repository::{
    HaveLivestreamCommentRepository, LivestreamCommentRepository,
};
use crate::services::ServiceResult;
use async_trait::async_trait;

#[async_trait]
pub trait LivestreamCommentService {
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
    async fn get_sum_tip(&self) -> ServiceResult<i64> {
        let mut conn = self.get_db_pool().acquire().await?;
        let sum_tip = self
            .livestream_comment_repo()
            .get_sum_tip(&mut *conn)
            .await?;
        Ok(sum_tip)
    }
}
