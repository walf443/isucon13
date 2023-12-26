use crate::models::id::Id;
use crate::models::livestream::LivestreamId;
use crate::models::user::UserId;

#[derive(Debug, sqlx::FromRow)]
pub struct Reaction {
    pub id: Id<Self>,
    pub emoji_name: String,
    pub user_id: UserId,
    pub livestream_id: LivestreamId,
    pub created_at: i64,
}

pub type ReactionId = Id<Reaction>;
