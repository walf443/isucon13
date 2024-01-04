use crate::repos::livestream_repository::LivestreamRepositoryInfra;
use crate::repos::livestream_tag_repository::LivestreamTagRepositoryInfra;
use crate::repos::tag_repository::TagRepositoryInfra;
use isupipe_core::db::{DBPool, HaveDBPool};
use isupipe_core::repos::livestream_repository::HaveLivestreamRepository;
use isupipe_core::repos::livestream_tag_repository::HaveLivestreamTagRepository;
use isupipe_core::repos::tag_repository::HaveTagRepository;
use isupipe_core::services::livestream_service::LivestreamServiceImpl;

#[derive(Clone)]
pub struct LivestreamServiceInfra {
    db_pool: DBPool,
    livestream_repo: LivestreamRepositoryInfra,
    livestream_tag_repo: LivestreamTagRepositoryInfra,
    tag_repo: TagRepositoryInfra,
}

impl LivestreamServiceInfra {
    pub fn new(db_pool: DBPool) -> Self {
        Self {
            db_pool,
            livestream_repo: LivestreamRepositoryInfra {},
            livestream_tag_repo: LivestreamTagRepositoryInfra {},
            tag_repo: TagRepositoryInfra {},
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

impl HaveLivestreamTagRepository for LivestreamServiceInfra {
    type Repo = LivestreamTagRepositoryInfra;

    fn livestream_tag_repo(&self) -> &Self::Repo {
        &self.livestream_tag_repo
    }
}

impl HaveTagRepository for LivestreamServiceInfra {
    type Repo = TagRepositoryInfra;

    fn tag_repo(&self) -> &Self::Repo {
        &self.tag_repo
    }
}

impl LivestreamServiceImpl for LivestreamServiceInfra {}
