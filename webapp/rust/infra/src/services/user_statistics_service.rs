use crate::repos::livestream_comment_repository::LivestreamCommentRepositoryInfra;
use crate::repos::livestream_repository::LivestreamRepositoryInfra;
use crate::repos::livestream_viewers_history_repository::LivestreamViewersHistoryRepositoryInfra;
use crate::repos::reaction_repository::ReactionRepositoryInfra;
use crate::repos::user_repository::UserRepositoryInfra;
use isupipe_core::db::{DBPool, HaveDBPool};
use isupipe_core::repos::livestream_comment_repository::HaveLivestreamCommentRepository;
use isupipe_core::repos::livestream_repository::HaveLivestreamRepository;
use isupipe_core::repos::livestream_viewers_history_repository::HaveLivestreamViewersHistoryRepository;
use isupipe_core::repos::reaction_repository::HaveReactionRepository;
use isupipe_core::repos::user_repository::HaveUserRepository;
use isupipe_core::services::user_statistics_service::UserStatisticsServiceImpl;
use std::sync::Arc;

pub struct UserStatisticsServiceInfra {
    db_pool: Arc<DBPool>,
    reaction_repo: ReactionRepositoryInfra,
    livestream_repo: LivestreamRepositoryInfra,
    livestream_comment_repo: LivestreamCommentRepositoryInfra,
    livestream_viewers_history_repo: LivestreamViewersHistoryRepositoryInfra,
    user_repo: UserRepositoryInfra,
}

impl UserStatisticsServiceInfra {
    pub fn new(db_pool: Arc<DBPool>) -> Self {
        Self {
            db_pool,
            reaction_repo: ReactionRepositoryInfra {},
            livestream_repo: LivestreamRepositoryInfra {},
            livestream_comment_repo: LivestreamCommentRepositoryInfra {},
            livestream_viewers_history_repo: LivestreamViewersHistoryRepositoryInfra {},
            user_repo: UserRepositoryInfra {},
        }
    }
}

impl HaveDBPool for UserStatisticsServiceInfra {
    fn get_db_pool(&self) -> &DBPool {
        &self.db_pool
    }
}

impl HaveReactionRepository for UserStatisticsServiceInfra {
    type Repo = ReactionRepositoryInfra;

    fn reaction_repo(&self) -> &Self::Repo {
        &self.reaction_repo
    }
}

impl HaveLivestreamCommentRepository for UserStatisticsServiceInfra {
    type Repo = LivestreamCommentRepositoryInfra;

    fn livestream_comment_repo(&self) -> &Self::Repo {
        &self.livestream_comment_repo
    }
}

impl HaveLivestreamRepository for UserStatisticsServiceInfra {
    type Repo = LivestreamRepositoryInfra;

    fn livestream_repo(&self) -> &Self::Repo {
        &self.livestream_repo
    }
}

impl HaveLivestreamViewersHistoryRepository for UserStatisticsServiceInfra {
    type Repo = LivestreamViewersHistoryRepositoryInfra;

    fn livestream_viewers_history_repo(&self) -> &Self::Repo {
        &self.livestream_viewers_history_repo
    }
}

impl HaveUserRepository for UserStatisticsServiceInfra {
    type Repo = UserRepositoryInfra;

    fn user_repo(&self) -> &Self::Repo {
        &self.user_repo
    }
}

impl UserStatisticsServiceImpl for UserStatisticsServiceInfra {}
