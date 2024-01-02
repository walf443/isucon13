use crate::repos::livestream_tag_repository::LivestreamTagRepositoryInfra;
use isupipe_core::db::{DBPool, HaveDBPool};
use isupipe_core::repos::livestream_tag_repository::HaveLivestreamTagRepository;
use isupipe_core::services::livestream_tag_service::LivestreamTagServiceImpl;

#[derive(Clone)]
pub struct LivestreamTagServiceInfra {
    db_pool: DBPool,
    livestream_tag_repo: LivestreamTagRepositoryInfra,
}

impl LivestreamTagServiceInfra {
    pub fn new(db_pool: DBPool) -> Self {
        Self {
            db_pool,
            livestream_tag_repo: LivestreamTagRepositoryInfra {},
        }
    }
}

impl HaveDBPool for LivestreamTagServiceInfra {
    fn get_db_pool(&self) -> &DBPool {
        &self.db_pool
    }
}

impl HaveLivestreamTagRepository for LivestreamTagServiceInfra {
    type Repo = LivestreamTagRepositoryInfra;

    fn livestream_tag_repo(&self) -> &Self::Repo {
        &self.livestream_tag_repo
    }
}

impl LivestreamTagServiceImpl for LivestreamTagServiceInfra {}
