use crate::services::livestream_comment_report_service::HaveLivestreamCommentReportService;
use crate::services::livestream_service::HaveLivestreamService;

pub trait ServiceManager: HaveLivestreamCommentReportService + HaveLivestreamService {}
