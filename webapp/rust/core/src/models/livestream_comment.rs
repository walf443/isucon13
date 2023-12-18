
#[derive(Debug, sqlx::FromRow)]
pub struct LivestreamCommentModel {
    pub id: i64,
    pub user_id: i64,
    pub livestream_id: i64,
    pub comment: String,
    pub tip: i64,
    pub created_at: i64,
}
