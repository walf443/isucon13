use crate::services::icon_service::IconServiceInfra;
use crate::services::initialize_service::InitializeServiceInfra;
use crate::services::livestream_comment_report_service::LivestreamCommentReportServiceInfra;
use crate::services::livestream_comment_service::LivestreamCommentServiceInfra;
use crate::services::livestream_service::LivestreamServiceInfra;
use crate::services::livestream_tag_service::LivestreamTagServiceInfra;
use crate::services::livestream_viewers_history_service::LivestreamViewersHistoryServiceInfra;
use crate::services::ng_word_service::NgWordServiceInfra;
use crate::services::reaction_service::ReactionServiceInfra;
use crate::services::tag_service::TagServiceInfra;
use crate::services::theme_service::ThemeServiceInfra;
use crate::services::user_service::UserServiceInfra;
use crate::services::user_statistics_service::UserStatisticsServiceInfra;
use isupipe_core::db::DBPool;
use isupipe_core::services::icon_service::HaveIconService;
use isupipe_core::services::initialize_service::HaveInitializeService;
use isupipe_core::services::livestream_comment_report_service::HaveLivestreamCommentReportService;
use isupipe_core::services::livestream_comment_service::HaveLivestreamCommentService;
use isupipe_core::services::livestream_service::HaveLivestreamService;
use isupipe_core::services::livestream_tag_service::HaveLivestreamTagService;
use isupipe_core::services::livestream_viewers_history_service::HaveLivestreamViewersHistoryService;
use isupipe_core::services::manager::ServiceManager;
use isupipe_core::services::ng_word_service::HaveNgWordService;
use isupipe_core::services::reaction_service::HaveReactionService;
use isupipe_core::services::tag_service::HaveTagService;
use isupipe_core::services::theme_service::HaveThemeService;
use isupipe_core::services::user_service::HaveUserService;
use isupipe_core::services::user_statistics_service::HaveUserStatisticsService;

#[derive(Clone)]
pub struct ServiceManagerInfra {
    icon_service: IconServiceInfra,
    initialize_service: InitializeServiceInfra,
    livestream_service: LivestreamServiceInfra,
    livestream_comment_service: LivestreamCommentServiceInfra,
    livestream_comment_report_service: LivestreamCommentReportServiceInfra,
    livestream_viewers_history_service: LivestreamViewersHistoryServiceInfra,
    livestream_tag_service: LivestreamTagServiceInfra,
    reaction_service: ReactionServiceInfra,
    ng_word_service: NgWordServiceInfra,
    tag_service: TagServiceInfra,
    user_service: UserServiceInfra,
    user_statistics_service: UserStatisticsServiceInfra,
    theme_service: ThemeServiceInfra,
}

impl ServiceManagerInfra {
    pub fn new(db_pool: DBPool) -> Self {
        Self {
            icon_service: IconServiceInfra::new(db_pool.clone()),
            initialize_service: InitializeServiceInfra::new(),
            livestream_service: LivestreamServiceInfra::new(db_pool.clone()),
            livestream_comment_service: LivestreamCommentServiceInfra::new(db_pool.clone()),
            livestream_comment_report_service: LivestreamCommentReportServiceInfra::new(
                db_pool.clone(),
            ),
            livestream_viewers_history_service: LivestreamViewersHistoryServiceInfra::new(
                db_pool.clone(),
            ),
            livestream_tag_service: LivestreamTagServiceInfra::new(db_pool.clone()),
            reaction_service: ReactionServiceInfra::new(db_pool.clone()),
            ng_word_service: NgWordServiceInfra::new(db_pool.clone()),
            tag_service: TagServiceInfra::new(db_pool.clone()),
            user_service: UserServiceInfra::new(db_pool.clone()),
            user_statistics_service: UserStatisticsServiceInfra::new(db_pool.clone()),
            theme_service: ThemeServiceInfra::new(db_pool.clone()),
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

impl HaveUserStatisticsService for ServiceManagerInfra {
    type Service = UserStatisticsServiceInfra;

    fn user_statistics_service(&self) -> &Self::Service {
        &self.user_statistics_service
    }
}

impl HaveThemeService for ServiceManagerInfra {
    type Service = ThemeServiceInfra;

    fn theme_service(&self) -> &Self::Service {
        &self.theme_service
    }
}

impl HaveNgWordService for ServiceManagerInfra {
    type Service = NgWordServiceInfra;

    fn ng_word_service(&self) -> &Self::Service {
        &self.ng_word_service
    }
}

impl HaveLivestreamViewersHistoryService for ServiceManagerInfra {
    type Service = LivestreamViewersHistoryServiceInfra;

    fn livestream_viewers_history_service(&self) -> &Self::Service {
        &self.livestream_viewers_history_service
    }
}

impl HaveLivestreamTagService for ServiceManagerInfra {
    type Service = LivestreamTagServiceInfra;

    fn livestream_tag_service(&self) -> &Self::Service {
        &self.livestream_tag_service
    }
}

impl HaveInitializeService for ServiceManagerInfra {
    type Service = InitializeServiceInfra;

    fn initialize_service(&self) -> &Self::Service {
        &self.initialize_service
    }
}

impl ServiceManager for ServiceManagerInfra {}
