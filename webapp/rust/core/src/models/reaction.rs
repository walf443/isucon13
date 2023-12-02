use crate::models::livestream::Livestream;
use crate::models::user::User;

#[derive(Debug, serde::Serialize)]
pub struct Reaction {
    pub id: i64,
    pub emoji_name: String,
    pub user: User,
    pub livestream: Livestream,
    pub created_at: i64,
}

#[derive(Debug, sqlx::FromRow)]
pub struct ReactionModel {
    pub id: i64,
    pub emoji_name: String,
    pub user_id: i64,
    pub livestream_id: i64,
    pub created_at: i64,
}
