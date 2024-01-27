use crate::models::livestream::LivestreamId;
use crate::models::user::UserId;
use kubetsu::Id;

#[derive(Debug, sqlx::FromRow)]
pub struct LivestreamComment {
    pub id: Id<Self, i64>,
    pub user_id: UserId,
    pub livestream_id: LivestreamId,
    pub comment: String,
    pub tip: i64,
    pub created_at: i64,
}

pub type LivestreamCommentId = Id<LivestreamComment, i64>;

pub struct CreateLivestreamComment {
    pub user_id: UserId,
    pub livestream_id: LivestreamId,
    pub comment: String,
    pub tip: i64,
    pub created_at: i64,
}
