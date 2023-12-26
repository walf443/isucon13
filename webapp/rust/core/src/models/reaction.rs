use crate::models::id::Id;

#[derive(Debug, sqlx::FromRow)]
pub struct Reaction {
    pub id: Id<Self>,
    pub emoji_name: String,
    pub user_id: i64,
    pub livestream_id: i64,
    pub created_at: i64,
}

pub type ReactionId = Id<Reaction>;
