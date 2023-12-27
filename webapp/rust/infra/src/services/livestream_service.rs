use crate::repos::livestream_repository::LivestreamRepositoryInfra;
use isupipe_core::db::{DBPool, HaveDBPool};
use isupipe_core::repos::livestream_repository::HaveLivestreamRepository;
use isupipe_core::services::livestream_service::LivestreamServiceImpl;
use std::sync::Arc;

pub struct LivestreamServiceInfra {
    db_pool: Arc<DBPool>,
    livestream_repo: LivestreamRepositoryInfra,
}

impl LivestreamServiceInfra {
    pub fn new(db_pool: Arc<DBPool>) -> Self {
        Self {
            db_pool,
            livestream_repo: LivestreamRepositoryInfra {},
        }
    }
}

impl HaveDBPool for LivestreamServiceInfra {
    fn get_db_pool(&self) -> &DBPool {
        &self.db_pool
    }
}

impl HaveLivestreamRepository for LivestreamServiceInfra {
    type Repo = LivestreamRepositoryInfra;

    fn livestream_repo(&self) -> &Self::Repo {
        &self.livestream_repo
    }
}

impl LivestreamServiceImpl for LivestreamServiceInfra {}
