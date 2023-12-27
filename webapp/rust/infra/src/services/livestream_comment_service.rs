use crate::repos::livestream_comment_repository::LivestreamCommentRepositoryInfra;
use isupipe_core::db::{DBPool, HaveDBPool};
use isupipe_core::repos::livestream_comment_repository::HaveLivestreamCommentRepository;
use isupipe_core::services::livestream_comment_service::LivestreamCommentServiceImpl;
use std::sync::Arc;

pub struct LivestreamCommentServiceInfra {
    db_pool: Arc<DBPool>,
    livestream_comment_repo: LivestreamCommentRepositoryInfra,
}

impl LivestreamCommentServiceInfra {
    pub fn new(db_pool: Arc<DBPool>) -> Self {
        Self {
            db_pool,
            livestream_comment_repo: LivestreamCommentRepositoryInfra {},
        }
    }
}

impl HaveLivestreamCommentRepository for LivestreamCommentServiceInfra {
    type Repo = LivestreamCommentRepositoryInfra;

    fn livestream_comment_repo(&self) -> &Self::Repo {
        &self.livestream_comment_repo
    }
}

impl HaveDBPool for LivestreamCommentServiceInfra {
    fn get_db_pool(&self) -> &DBPool {
        &self.db_pool
    }
}

impl LivestreamCommentServiceImpl for LivestreamCommentServiceInfra {}
