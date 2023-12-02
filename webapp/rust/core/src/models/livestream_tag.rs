#[derive(Debug, sqlx::FromRow)]
pub struct LivestreamTagModel {
    #[allow(unused)]
    pub id: i64,
    pub livestream_id: i64,
    pub tag_id: i64,
}
