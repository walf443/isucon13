use crate::services::livestream_comment_report_service::HaveLivestreamCommentReportService;
use crate::services::livestream_comment_service::HaveLivestreamCommentService;
use crate::services::livestream_service::HaveLivestreamService;
use crate::services::reaction_service::HaveReactionService;
use crate::services::tag_service::HaveTagService;
use crate::services::user_icon_service::HaveUserIconService;

pub trait ServiceManager:
    HaveLivestreamCommentReportService
    + HaveLivestreamService
    + HaveReactionService
    + HaveLivestreamCommentService
    + HaveTagService
    + HaveUserIconService
{
}
