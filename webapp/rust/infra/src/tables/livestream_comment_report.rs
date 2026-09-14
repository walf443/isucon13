use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::livestream_comment::LivestreamCommentId;
use isupipe_core::models::livestream_comment_report::LivestreamCommentReportId;
use isupipe_core::models::user::UserId;

qbey::qbey_schema!(
    LivestreamCommentReportTable,
    "livecomment_reports",
    [
        id: LivestreamCommentReportId,
        user_id: UserId,
        livestream_id: LivestreamId,
        livecomment_id: LivestreamCommentId,
        created_at: i64,
    ]
);

pub const TABLE_LIVECOMMENT_REPORTS: LivestreamCommentReportTable =
    LivestreamCommentReportTable::new();
