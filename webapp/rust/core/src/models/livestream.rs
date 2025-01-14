use crate::models::user::UserId;
use kubetsu::Id;

pub struct CreateLivestream {
    pub user_id: UserId,
    pub title: String,
    pub description: String,
    pub playlist_url: String,
    pub thumbnail_url: String,
    pub start_at: i64,
    pub end_at: i64,
}

#[derive(Debug, sqlx::FromRow)]
pub struct Livestream {
    pub id: Id<Self, i64>,
    pub user_id: UserId,
    pub title: String,
    pub description: String,
    pub playlist_url: String,
    pub thumbnail_url: String,
    pub start_at: i64,
    pub end_at: i64,
}

pub type LivestreamId = Id<Livestream, i64>;
