use crate::models::id::Id;

#[derive(Debug, sqlx::FromRow)]
pub struct LivestreamComment {
    pub id: Id<Self>,
    pub user_id: i64,
    pub livestream_id: i64,
    pub comment: String,
    pub tip: i64,
    pub created_at: i64,
}

pub type LivestreamCommentId = Id<LivestreamComment>;
