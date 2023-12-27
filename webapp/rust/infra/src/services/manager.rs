use crate::services::icon_service::IconServiceInfra;
use crate::services::livestream_comment_report_service::LivestreamCommentReportServiceInfra;
use crate::services::livestream_comment_service::LivestreamCommentServiceInfra;
use crate::services::livestream_service::LivestreamServiceInfra;
use crate::services::reaction_service::ReactionServiceInfra;
use crate::services::tag_service::TagServiceInfra;
use crate::services::user_service::UserServiceInfra;
use isupipe_core::db::DBPool;
use isupipe_core::services::icon_service::HaveIconService;
use isupipe_core::services::livestream_comment_report_service::HaveLivestreamCommentReportService;
use isupipe_core::services::livestream_comment_service::HaveLivestreamCommentService;
use isupipe_core::services::livestream_service::HaveLivestreamService;
use isupipe_core::services::manager::ServiceManager;
use isupipe_core::services::reaction_service::HaveReactionService;
use isupipe_core::services::tag_service::HaveTagService;
use isupipe_core::services::user_service::HaveUserService;
use std::sync::Arc;

pub struct ServiceManagerInfra {
    icon_service: IconServiceInfra,
    livestream_service: LivestreamServiceInfra,
    livestream_comment_service: LivestreamCommentServiceInfra,
    livestream_comment_report_service: LivestreamCommentReportServiceInfra,
    reaction_service: ReactionServiceInfra,
    tag_service: TagServiceInfra,
    user_service: UserServiceInfra,
}

impl ServiceManagerInfra {
    pub fn new(db_pool: DBPool) -> Self {
        let db_pool = Arc::new(db_pool);
        Self {
            icon_service: IconServiceInfra::new(db_pool.clone()),
            livestream_service: LivestreamServiceInfra::new(db_pool.clone()),
            livestream_comment_service: LivestreamCommentServiceInfra::new(db_pool.clone()),
            livestream_comment_report_service: LivestreamCommentReportServiceInfra::new(
                db_pool.clone(),
            ),
            reaction_service: ReactionServiceInfra::new(db_pool.clone()),
            tag_service: TagServiceInfra::new(db_pool.clone()),
            user_service: UserServiceInfra::new(db_pool.clone()),
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

impl HaveIconService for ServiceManagerInfra {
    type Service = IconServiceInfra;

    fn icon_service(&self) -> &Self::Service {
        &self.icon_service
    }
}

impl HaveUserService for ServiceManagerInfra {
    type Service = UserServiceInfra;

    fn user_service(&self) -> &Self::Service {
        &self.user_service
    }
}

impl ServiceManager for ServiceManagerInfra {}
