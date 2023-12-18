#[derive(Debug, sqlx::FromRow)]
pub struct ReactionModel {
    pub id: i64,
    pub emoji_name: String,
    pub user_id: i64,
    pub livestream_id: i64,
    pub created_at: i64,
}
