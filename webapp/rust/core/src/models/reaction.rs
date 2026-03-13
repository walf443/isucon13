use crate::models::livestream::LivestreamId;
use crate::models::user::UserId;
use fake::Dummy;
use kubetsu::Id;

pub const TABLE_NAME: &str = "reactions";

#[derive(Debug, sqlx::FromRow)]
pub struct Reaction {
    pub id: Id<Self, i64>,
    pub emoji_name: String,
    pub user_id: UserId,
    pub livestream_id: LivestreamId,
    pub created_at: i64,
}

#[derive(Dummy)]
pub struct CreateReaction {
    pub emoji_name: String,
    pub user_id: UserId,
    pub livestream_id: LivestreamId,
    pub created_at: i64,
}

pub type ReactionId = Id<Reaction, i64>;
