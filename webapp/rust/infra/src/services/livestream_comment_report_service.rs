use crate::repos::livestream_comment_report_repository::LivestreamCommentReportRepositoryInfra;
use crate::repos::livestream_comment_repository::LivestreamCommentRepositoryInfra;
use crate::repos::livestream_repository::LivestreamRepositoryInfra;
use isupipe_core::db::{DBPool, HaveDBPool};
use isupipe_core::repos::livestream_comment_report_repository::HaveLivestreamCommentReportRepository;
use isupipe_core::repos::livestream_comment_repository::HaveLivestreamCommentRepository;
use isupipe_core::repos::livestream_repository::HaveLivestreamRepository;
use isupipe_core::services::livestream_comment_report_service::LivestreamCommentReportServiceImpl;
use std::sync::Arc;

#[derive(Clone)]
pub struct LivestreamCommentReportServiceInfra {
    db_pool: Arc<DBPool>,
    livestream_repo: LivestreamRepositoryInfra,
    livestream_comment_repo: LivestreamCommentRepositoryInfra,
    livestream_comment_report_repo: LivestreamCommentReportRepositoryInfra,
}

impl LivestreamCommentReportServiceInfra {
    pub fn new(db_pool: Arc<DBPool>) -> Self {
        Self {
            db_pool,
            livestream_repo: LivestreamRepositoryInfra {},
            livestream_comment_repo: LivestreamCommentRepositoryInfra {},
            livestream_comment_report_repo: LivestreamCommentReportRepositoryInfra {},
        }
    }
}

impl LivestreamCommentReportServiceImpl for LivestreamCommentReportServiceInfra {}

impl HaveDBPool for LivestreamCommentReportServiceInfra {
    fn get_db_pool(&self) -> &DBPool {
        &self.db_pool
    }
}

impl HaveLivestreamCommentReportRepository for LivestreamCommentReportServiceInfra {
    type Repo = LivestreamCommentReportRepositoryInfra;

    fn livestream_comment_report_repo(&self) -> &Self::Repo {
        &self.livestream_comment_report_repo
    }
}

impl HaveLivestreamRepository for LivestreamCommentReportServiceInfra {
    type Repo = LivestreamRepositoryInfra;

    fn livestream_repo(&self) -> &Self::Repo {
        &self.livestream_repo
    }
}

impl HaveLivestreamCommentRepository for LivestreamCommentReportServiceInfra {
    type Repo = LivestreamCommentRepositoryInfra;

    fn livestream_comment_repo(&self) -> &Self::Repo {
        &self.livestream_comment_repo
    }
}
