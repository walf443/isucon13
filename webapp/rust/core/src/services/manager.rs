use crate::services::icon_service::HaveIconService;
use crate::services::livestream_comment_report_service::HaveLivestreamCommentReportService;
use crate::services::livestream_comment_service::HaveLivestreamCommentService;
use crate::services::livestream_service::HaveLivestreamService;
use crate::services::reaction_service::HaveReactionService;
use crate::services::tag_service::HaveTagService;
use crate::services::theme_service::HaveThemeService;
use crate::services::user_service::HaveUserService;
use crate::services::user_statistics_service::HaveUserStatisticsService;

pub trait ServiceManager:
    HaveLivestreamCommentReportService
    + HaveLivestreamService
    + HaveReactionService
    + HaveLivestreamCommentService
    + HaveTagService
    + HaveIconService
    + HaveUserService
    + HaveUserStatisticsService
    + HaveThemeService
{
}
