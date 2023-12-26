use crate::models::id::Id;
use crate::models::livestream::LivestreamId;
use crate::models::livestream_comment::LivestreamCommentId;
use crate::models::user::UserId;

#[derive(Debug, sqlx::FromRow)]
pub struct LivestreamCommentReport {
    pub id: Id<Self>,
    pub user_id: UserId,
    #[allow(unused)]
    pub livestream_id: LivestreamId,
    pub livestream_comment_id: LivestreamCommentId,
    pub created_at: i64,
}

pub struct CreateLivestreamCommentReport {
    pub user_id: UserId,
    #[allow(unused)]
    pub livestream_id: LivestreamId,
    pub livestream_comment_id: LivestreamCommentId,
    pub created_at: i64,
}

pub type LivestreamCommentReportId = Id<LivestreamCommentReport>;
