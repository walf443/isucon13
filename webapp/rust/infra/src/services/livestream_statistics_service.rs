use crate::repos::livestream_comment_report_repository::LivestreamCommentReportRepositoryInfra;
use crate::repos::livestream_comment_repository::LivestreamCommentRepositoryInfra;
use crate::repos::livestream_repository::LivestreamRepositoryInfra;
use crate::repos::livestream_viewers_history_repository::LivestreamViewersHistoryRepositoryInfra;
use crate::repos::reaction_repository::ReactionRepositoryInfra;
use isupipe_core::db::{DBPool, HaveDBPool};
use isupipe_core::repos::livestream_comment_report_repository::HaveLivestreamCommentReportRepository;
use isupipe_core::repos::livestream_comment_repository::HaveLivestreamCommentRepository;
use isupipe_core::repos::livestream_repository::HaveLivestreamRepository;
use isupipe_core::repos::livestream_viewers_history_repository::HaveLivestreamViewersHistoryRepository;
use isupipe_core::repos::reaction_repository::HaveReactionRepository;
use isupipe_core::services::livestream_statistics_service::LivestreamStatisticsServiceImpl;

#[derive(Clone)]
pub struct LivestreamStatisticsServiceInfra {
    db_pool: DBPool,
    livestream_comment_repo: LivestreamCommentRepositoryInfra,
    livestream_comment_report_repo: LivestreamCommentReportRepositoryInfra,
    livestream_repo: LivestreamRepositoryInfra,
    livestream_viewers_history_repo: LivestreamViewersHistoryRepositoryInfra,
    reaction_repo: ReactionRepositoryInfra,
}

impl LivestreamStatisticsServiceInfra {
    pub fn new(db_pool: DBPool) -> Self {
        Self {
            db_pool,
            livestream_comment_repo: LivestreamCommentRepositoryInfra {},
            livestream_comment_report_repo: LivestreamCommentReportRepositoryInfra {},
            livestream_repo: LivestreamRepositoryInfra {},
            livestream_viewers_history_repo: LivestreamViewersHistoryRepositoryInfra {},
            reaction_repo: ReactionRepositoryInfra {},
        }
    }
}

impl HaveDBPool for LivestreamStatisticsServiceInfra {
    fn get_db_pool(&self) -> &DBPool {
        &self.db_pool
    }
}

impl HaveLivestreamCommentReportRepository for LivestreamStatisticsServiceInfra {
    type Repo = LivestreamCommentReportRepositoryInfra;

    fn livestream_comment_report_repo(&self) -> &Self::Repo {
        &self.livestream_comment_report_repo
    }
}

impl HaveReactionRepository for LivestreamStatisticsServiceInfra {
    type Repo = ReactionRepositoryInfra;

    fn reaction_repo(&self) -> &Self::Repo {
        &self.reaction_repo
    }
}

impl HaveLivestreamCommentRepository for LivestreamStatisticsServiceInfra {
    type Repo = LivestreamCommentRepositoryInfra;

    fn livestream_comment_repo(&self) -> &Self::Repo {
        &self.livestream_comment_repo
    }
}

impl HaveLivestreamViewersHistoryRepository for LivestreamStatisticsServiceInfra {
    type Repo = LivestreamViewersHistoryRepositoryInfra;

    fn livestream_viewers_history_repo(&self) -> &Self::Repo {
        &self.livestream_viewers_history_repo
    }
}

impl HaveLivestreamRepository for LivestreamStatisticsServiceInfra {
    type Repo = LivestreamRepositoryInfra;

    fn livestream_repo(&self) -> &Self::Repo {
        &self.livestream_repo
    }
}

impl LivestreamStatisticsServiceImpl for LivestreamStatisticsServiceInfra {}
