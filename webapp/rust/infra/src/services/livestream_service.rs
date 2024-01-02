use crate::repos::livestream_repository::LivestreamRepositoryInfra;
use isupipe_core::db::{DBPool, HaveDBPool};
use isupipe_core::repos::livestream_repository::HaveLivestreamRepository;
use isupipe_core::services::livestream_service::LivestreamServiceImpl;

#[derive(Clone)]
pub struct LivestreamServiceInfra {
    db_pool: DBPool,
    livestream_repo: LivestreamRepositoryInfra,
}

impl LivestreamServiceInfra {
    pub fn new(db_pool: DBPool) -> Self {
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
