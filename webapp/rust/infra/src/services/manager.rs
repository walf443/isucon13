use crate::services::livestream_comment_report_service::LivestreamCommentReportServiceInfra;
use isupipe_core::db::DBPool;
use isupipe_core::services::livestream_comment_report_service::HaveLivestreamCommentReportService;
use isupipe_core::services::manager::ServiceManager;
use std::sync::Arc;

pub struct ServiceManagerInfra {
    livestream_comment_report_service: LivestreamCommentReportServiceInfra,
}

impl ServiceManagerInfra {
    pub fn new(db_pool: DBPool) -> Self {
        let db_pool = Arc::new(db_pool);
        Self {
            livestream_comment_report_service: LivestreamCommentReportServiceInfra::new(
                db_pool.clone(),
            ),
        }
    }
}

impl HaveLivestreamCommentReportService for ServiceManagerInfra {
    type Service = LivestreamCommentReportServiceInfra;

    fn livestream_comment_report_service(&self) -> &Self::Service {
        &self.livestream_comment_report_service
    }
}

impl ServiceManager for ServiceManagerInfra {}
