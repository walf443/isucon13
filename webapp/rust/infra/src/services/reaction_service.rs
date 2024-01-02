use crate::repos::reaction_repository::ReactionRepositoryInfra;
use isupipe_core::db::{DBPool, HaveDBPool};
use isupipe_core::repos::reaction_repository::HaveReactionRepository;
use isupipe_core::services::reaction_service::ReactionServiceImpl;
use std::sync::Arc;

#[derive(Clone)]
pub struct ReactionServiceInfra {
    db_pool: Arc<DBPool>,
    reaction_repo: ReactionRepositoryInfra,
}

impl ReactionServiceInfra {
    pub fn new(db_pool: Arc<DBPool>) -> Self {
        Self {
            db_pool,
            reaction_repo: ReactionRepositoryInfra {},
        }
    }
}

impl HaveDBPool for ReactionServiceInfra {
    fn get_db_pool(&self) -> &DBPool {
        &self.db_pool
    }
}

impl HaveReactionRepository for ReactionServiceInfra {
    type Repo = ReactionRepositoryInfra;

    fn reaction_repo(&self) -> &Self::Repo {
        &self.reaction_repo
    }
}

impl ReactionServiceImpl for ReactionServiceInfra {}
