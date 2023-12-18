#[derive(Debug, sqlx::FromRow)]
pub struct LivestreamTag {
    #[allow(unused)]
    pub id: i64,
    pub livestream_id: i64,
    pub tag_id: i64,
}
