
#[derive(Debug, serde::Serialize, sqlx::FromRow)]
pub struct NgWord {
    pub id: i64,
    pub user_id: i64,
    pub livestream_id: i64,
    pub word: String,
    #[sqlx(default)]
    pub created_at: i64,
}
