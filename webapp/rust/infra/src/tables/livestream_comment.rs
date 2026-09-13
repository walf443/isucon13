use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::livestream_comment::{LivestreamComment, LivestreamCommentId};
use isupipe_core::models::user::UserId;

#[derive(Debug, toasty::Model)]
#[table = "livecomments"]
pub struct LivestreamCommentRow {
    #[key]
    #[auto]
    pub id: LivestreamCommentId,
    pub user_id: UserId,
    pub livestream_id: LivestreamId,
    pub comment: String,
    pub tip: i64,
    pub created_at: i64,
}

impl From<LivestreamCommentRow> for LivestreamComment {
    fn from(row: LivestreamCommentRow) -> Self {
        Self {
            id: row.id,
            user_id: row.user_id,
            livestream_id: row.livestream_id,
            comment: row.comment,
            tip: row.tip,
            created_at: row.created_at,
        }
    }
}
