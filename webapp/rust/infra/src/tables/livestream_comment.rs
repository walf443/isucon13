use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::livestream_comment::{LivestreamComment, LivestreamCommentId};
use isupipe_core::models::user::UserId;

#[derive(Debug, toasty::Model)]
#[table = "livecomments"]
pub struct LivestreamCommentRow {
    #[key]
    #[auto]
    pub id: i64,
    pub user_id: i64,
    pub livestream_id: i64,
    pub comment: String,
    pub tip: i64,
    pub created_at: i64,
}

impl From<LivestreamCommentRow> for LivestreamComment {
    fn from(row: LivestreamCommentRow) -> Self {
        Self {
            id: LivestreamCommentId::new(row.id),
            user_id: UserId::new(row.user_id),
            livestream_id: LivestreamId::new(row.livestream_id),
            comment: row.comment,
            tip: row.tip,
            created_at: row.created_at,
        }
    }
}
