use crate::services::livestream_comment_report_service::LivestreamCommentReportServiceInfra;
use crate::services::livestream_service::LivestreamServiceInfra;
use crate::services::reaction_service::ReactionServiceInfra;
use isupipe_core::db::DBPool;
use isupipe_core::services::livestream_comment_report_service::HaveLivestreamCommentReportService;
use isupipe_core::services::livestream_service::HaveLivestreamService;
use isupipe_core::services::manager::ServiceManager;
use isupipe_core::services::reaction_service::HaveReactionService;
use std::sync::Arc;

pub struct ServiceManagerInfra {
    livestream_service: LivestreamServiceInfra,
    livestream_comment_report_service: LivestreamCommentReportServiceInfra,
    reaction_service: ReactionServiceInfra,
}

impl ServiceManagerInfra {
    pub fn new(db_pool: DBPool) -> Self {
        let db_pool = Arc::new(db_pool);
        Self {
            livestream_service: LivestreamServiceInfra::new(db_pool.clone()),
            livestream_comment_report_service: LivestreamCommentReportServiceInfra::new(
                db_pool.clone(),
            ),
            reaction_service: ReactionServiceInfra::new(db_pool.clone()),
        }
    }
}

impl HaveLivestreamCommentReportService for ServiceManagerInfra {
    type Service = LivestreamCommentReportServiceInfra;

    fn livestream_comment_report_service(&self) -> &Self::Service {
        &self.livestream_comment_report_service
    }
}

impl HaveLivestreamService for ServiceManagerInfra {
    type Service = LivestreamServiceInfra;

    fn livestream_service(&self) -> &Self::Service {
        &self.livestream_service
    }
}

impl HaveReactionService for ServiceManagerInfra {
    type Service = ReactionServiceInfra;

    fn reaction_service(&self) -> &Self::Service {
        &self.reaction_service
    }
}

impl ServiceManager for ServiceManagerInfra {}
