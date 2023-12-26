use crate::db::HaveDBPool;
use crate::repos::livestream_comment_report_repository::HaveLivestreamCommentReportRepository;

pub trait RepositoryManager: HaveDBPool + HaveLivestreamCommentReportRepository {}
