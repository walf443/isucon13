use crate::models::id::Id;
use crate::models::livestream_comment::LivestreamCommentId;

#[derive(Debug, sqlx::FromRow)]
pub struct LivestreamCommentReport {
    pub id: Id<Self>,
    pub user_id: i64,
    #[allow(unused)]
    pub livestream_id: i64,
    pub livestream_comment_id: LivestreamCommentId,
    pub created_at: i64,
}

pub type LivestreamCommentReportId = Id<LivestreamCommentReport>;
