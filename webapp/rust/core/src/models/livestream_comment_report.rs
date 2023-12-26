use crate::models::id::Id;

#[derive(Debug, sqlx::FromRow)]
pub struct LivestreamCommentReport {
    pub id: Id<Self>,
    pub user_id: i64,
    #[allow(unused)]
    pub livestream_id: i64,
    pub livestream_comment_id: i64,
    pub created_at: i64,
}

pub type LivestreamCommentReportId = Id<LivestreamCommentReport>;
