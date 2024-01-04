use crate::models::id::Id;
use crate::models::livestream::LivestreamId;
use crate::models::user::UserId;

#[derive(Debug, sqlx::FromRow)]
pub struct LivestreamComment {
    pub id: Id<Self>,
    pub user_id: UserId,
    pub livestream_id: LivestreamId,
    pub comment: String,
    pub tip: i64,
    pub created_at: i64,
}

pub type LivestreamCommentId = Id<LivestreamComment>;

pub struct CreateLivestreamComment {
    pub user_id: UserId,
    pub livestream_id: LivestreamId,
    pub comment: String,
    pub tip: i64,
    pub created_at: i64,
}
