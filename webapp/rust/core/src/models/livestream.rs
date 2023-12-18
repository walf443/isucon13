pub struct CreateLivestreamModel {
    pub user_id: i64,
    pub title: String,
    pub description: String,
    pub playlist_url: String,
    pub thumbnail_url: String,
    pub start_at: i64,
    pub end_at: i64,
}

#[derive(Debug, sqlx::FromRow)]
pub struct LivestreamModel {
    pub id: i64,
    pub user_id: i64,
    pub title: String,
    pub description: String,
    pub playlist_url: String,
    pub thumbnail_url: String,
    pub start_at: i64,
    pub end_at: i64,
}
