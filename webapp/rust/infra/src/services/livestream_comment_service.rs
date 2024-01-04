use crate::repos::livestream_comment_repository::LivestreamCommentRepositoryInfra;
use crate::repos::ng_word_repository::NgWordRepositoryInfra;
use isupipe_core::db::{DBPool, HaveDBPool};
use isupipe_core::repos::livestream_comment_repository::HaveLivestreamCommentRepository;
use isupipe_core::repos::ng_word_repository::HaveNgWordRepository;
use isupipe_core::services::livestream_comment_service::LivestreamCommentServiceImpl;

#[derive(Clone)]
pub struct LivestreamCommentServiceInfra {
    db_pool: DBPool,
    livestream_comment_repo: LivestreamCommentRepositoryInfra,
    ng_word_repo: NgWordRepositoryInfra,
}

impl LivestreamCommentServiceInfra {
    pub fn new(db_pool: DBPool) -> Self {
        Self {
            db_pool,
            livestream_comment_repo: LivestreamCommentRepositoryInfra {},
            ng_word_repo: NgWordRepositoryInfra {},
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

impl HaveNgWordRepository for LivestreamCommentServiceInfra {
    type Repo = NgWordRepositoryInfra;

    fn ng_word_repo(&self) -> &Self::Repo {
        &self.ng_word_repo
    }
}

impl LivestreamCommentServiceImpl for LivestreamCommentServiceInfra {}
