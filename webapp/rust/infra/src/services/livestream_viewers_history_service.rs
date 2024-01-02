use crate::repos::livestream_viewers_history_repository::LivestreamViewersHistoryRepositoryInfra;
use isupipe_core::db::{DBPool, HaveDBPool};
use isupipe_core::repos::livestream_viewers_history_repository::HaveLivestreamViewersHistoryRepository;
use isupipe_core::services::livestream_viewers_history_service::LivestreamViewersHistoryServiceImpl;
use std::sync::Arc;

#[derive(Clone)]
pub struct LivestreamViewersHistoryServiceInfra {
    db_pool: Arc<DBPool>,
    livestream_viewers_history_repo: LivestreamViewersHistoryRepositoryInfra,
}

impl LivestreamViewersHistoryServiceInfra {
    pub fn new(db_pool: Arc<DBPool>) -> Self {
        Self {
            db_pool,
            livestream_viewers_history_repo: LivestreamViewersHistoryRepositoryInfra {},
        }
    }
}

impl HaveDBPool for LivestreamViewersHistoryServiceInfra {
    fn get_db_pool(&self) -> &DBPool {
        &self.db_pool
    }
}

impl HaveLivestreamViewersHistoryRepository for LivestreamViewersHistoryServiceInfra {
    type Repo = LivestreamViewersHistoryRepositoryInfra;

    fn livestream_viewers_history_repo(&self) -> &Self::Repo {
        &self.livestream_viewers_history_repo
    }
}

impl LivestreamViewersHistoryServiceImpl for LivestreamViewersHistoryServiceInfra {}
