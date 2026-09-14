use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::livestream_comment::LivestreamCommentId;
use isupipe_core::models::livestream_comment_report::{
    LivestreamCommentReport, LivestreamCommentReportId,
};
use isupipe_core::models::user::UserId;

#[derive(Debug, toasty::Model)]
#[table = "livecomment_reports"]
pub struct LivestreamCommentReportRow {
    #[key]
    #[auto]
    pub id: LivestreamCommentReportId,
    pub user_id: UserId,
    pub livestream_id: LivestreamId,
    pub livecomment_id: LivestreamCommentId,
    pub created_at: i64,
}

impl From<LivestreamCommentReportRow> for LivestreamCommentReport {
    fn from(row: LivestreamCommentReportRow) -> Self {
        Self {
            id: row.id,
            user_id: row.user_id,
            livestream_id: row.livestream_id,
            livestream_comment_id: row.livecomment_id,
            created_at: row.created_at,
        }
    }
}
