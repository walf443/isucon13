use crate::repos::ng_word_repository::NgWordRepositoryInfra;
use isupipe_core::db::{DBPool, HaveDBPool};
use isupipe_core::repos::ng_word_repository::HaveNgWordRepository;
use isupipe_core::services::ng_word_service::NgWordServiceImpl;
use std::sync::Arc;

#[derive(Clone)]
pub struct NgWordServiceInfra {
    db_pool: Arc<DBPool>,
    ng_word_repo: NgWordRepositoryInfra,
}

impl NgWordServiceInfra {
    pub fn new(db_pool: Arc<DBPool>) -> Self {
        Self {
            db_pool,
            ng_word_repo: NgWordRepositoryInfra {},
        }
    }
}

impl HaveDBPool for NgWordServiceInfra {
    fn get_db_pool(&self) -> &DBPool {
        &self.db_pool
    }
}

impl HaveNgWordRepository for NgWordServiceInfra {
    type Repo = NgWordRepositoryInfra;

    fn ng_word_repo(&self) -> &Self::Repo {
        &self.ng_word_repo
    }
}

impl NgWordServiceImpl for NgWordServiceInfra {}
