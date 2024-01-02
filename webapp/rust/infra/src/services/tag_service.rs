use crate::repos::tag_repository::TagRepositoryInfra;
use isupipe_core::db::{DBPool, HaveDBPool};
use isupipe_core::repos::tag_repository::HaveTagRepository;
use isupipe_core::services::tag_service::TagServiceImpl;
use std::sync::Arc;

#[derive(Clone)]
pub struct TagServiceInfra {
    db_pool: Arc<DBPool>,
    tag_repo: TagRepositoryInfra,
}

impl TagServiceInfra {
    pub fn new(db_pool: Arc<DBPool>) -> Self {
        Self {
            db_pool,
            tag_repo: TagRepositoryInfra {},
        }
    }
}

impl HaveDBPool for TagServiceInfra {
    fn get_db_pool(&self) -> &DBPool {
        &self.db_pool
    }
}

impl HaveTagRepository for TagServiceInfra {
    type Repo = TagRepositoryInfra;

    fn tag_repo(&self) -> &Self::Repo {
        &self.tag_repo
    }
}

impl TagServiceImpl for TagServiceInfra {}
