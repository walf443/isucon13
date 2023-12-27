use crate::services::livestream_comment_report_service::LivestreamCommentReportServiceInfra;
use crate::services::livestream_comment_service::LivestreamCommentServiceInfra;
use crate::services::livestream_service::LivestreamServiceInfra;
use crate::services::reaction_service::ReactionServiceInfra;
use crate::services::tag_service::TagServiceInfra;
use crate::services::user_icon_service::UserIconServiceInfra;
use isupipe_core::db::DBPool;
use isupipe_core::services::livestream_comment_report_service::HaveLivestreamCommentReportService;
use isupipe_core::services::livestream_comment_service::HaveLivestreamCommentService;
use isupipe_core::services::livestream_service::HaveLivestreamService;
use isupipe_core::services::manager::ServiceManager;
use isupipe_core::services::reaction_service::HaveReactionService;
use isupipe_core::services::tag_service::HaveTagService;
use isupipe_core::services::user_icon_service::HaveUserIconService;
use std::sync::Arc;

pub struct ServiceManagerInfra {
    livestream_service: LivestreamServiceInfra,
    livestream_comment_service: LivestreamCommentServiceInfra,
    livestream_comment_report_service: LivestreamCommentReportServiceInfra,
    reaction_service: ReactionServiceInfra,
    tag_service: TagServiceInfra,
    user_icon_service: UserIconServiceInfra,
}

impl ServiceManagerInfra {
    pub fn new(db_pool: DBPool) -> Self {
        let db_pool = Arc::new(db_pool);
        Self {
            livestream_service: LivestreamServiceInfra::new(db_pool.clone()),
            livestream_comment_service: LivestreamCommentServiceInfra::new(db_pool.clone()),
            livestream_comment_report_service: LivestreamCommentReportServiceInfra::new(
                db_pool.clone(),
            ),
            reaction_service: ReactionServiceInfra::new(db_pool.clone()),
            tag_service: TagServiceInfra::new(db_pool.clone()),
            user_icon_service: UserIconServiceInfra::new(db_pool.clone()),
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

impl HaveLivestreamCommentService for ServiceManagerInfra {
    type Service = LivestreamCommentServiceInfra;

    fn livestream_comment_service(&self) -> &Self::Service {
        &self.livestream_comment_service
    }
}

impl HaveTagService for ServiceManagerInfra {
    type Service = TagServiceInfra;

    fn tag_service(&self) -> &Self::Service {
        &self.tag_service
    }
}

impl HaveUserIconService for ServiceManagerInfra {
    type Service = UserIconServiceInfra;

    fn user_icon_service(&self) -> &Self::Service {
        &self.user_icon_service
    }
}

impl ServiceManager for ServiceManagerInfra {}
