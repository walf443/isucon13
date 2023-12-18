#[derive(Debug, sqlx::FromRow)]
pub struct LivestreamCommentReport {
    pub id: i64,
    pub user_id: i64,
    #[allow(unused)]
    pub livestream_id: i64,
    pub livecomment_id: i64,
    pub created_at: i64,
}
