use crate::db::HaveDBPool;
use crate::models::livestream::{Livestream, LivestreamId};
use crate::repos::livestream_repository::{HaveLivestreamRepository, LivestreamRepository};
use crate::services::ServiceResult;
use async_trait::async_trait;

#[async_trait]
pub trait LivestreamService {
    async fn find(&self, livestream_id: &LivestreamId) -> ServiceResult<Option<Livestream>>;
}

pub trait HaveLivestreamService {
    type Service: LivestreamService;

    fn livestream_service(&self) -> &Self::Service;
}

pub trait LivestreamServiceImpl: Sync + HaveDBPool + HaveLivestreamRepository {}

#[async_trait]
impl<T: LivestreamServiceImpl> LivestreamService for T {
    async fn find(&self, livestream_id: &LivestreamId) -> ServiceResult<Option<Livestream>> {
        let mut conn = self.get_db_pool().acquire().await?;
        let result = self
            .livestream_repo()
            .find(&mut *conn, livestream_id)
            .await?;
        Ok(result)
    }
}
