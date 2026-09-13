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
    pub id: i64,
    pub user_id: i64,
    pub livestream_id: i64,
    pub livecomment_id: i64,
    pub created_at: i64,
}

impl From<LivestreamCommentReportRow> for LivestreamCommentReport {
    fn from(row: LivestreamCommentReportRow) -> Self {
        Self {
            id: LivestreamCommentReportId::new(row.id),
            user_id: UserId::new(row.user_id),
            livestream_id: LivestreamId::new(row.livestream_id),
            livestream_comment_id: LivestreamCommentId::new(row.livecomment_id),
            created_at: row.created_at,
        }
    }
}
