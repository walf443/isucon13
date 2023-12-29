use crate::db::HaveDBPool;
use crate::models::livestream_viewers_history::CreateLivestreamViewersHistory;
use crate::repos::livestream_viewers_history_repository::{
    HaveLivestreamViewersHistoryRepository, LivestreamViewersHistoryRepository,
};
use crate::services::ServiceResult;
use async_trait::async_trait;

#[async_trait]
pub trait LivestreamViewersHistoryService {
    async fn create(&self, history: &CreateLivestreamViewersHistory) -> ServiceResult<()>;
}

pub trait HaveLivestreamViewersHistoryService {
    type Service: LivestreamViewersHistoryService;

    fn livestream_viewers_history_service(&self) -> &Self::Service;
}

pub trait LivestreamViewersHistoryServiceImpl:
    Sync + HaveDBPool + HaveLivestreamViewersHistoryRepository
{
}

#[async_trait]
impl<T: LivestreamViewersHistoryServiceImpl> LivestreamViewersHistoryService for T {
    async fn create(&self, history: &CreateLivestreamViewersHistory) -> ServiceResult<()> {
        let mut conn = self.get_db_pool().acquire().await?;
        self.livestream_viewers_history_repo()
            .create(&mut conn, history)
            .await?;
        Ok(())
    }
}
