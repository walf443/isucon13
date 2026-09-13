use isupipe_core::models::livestream::{Livestream, LivestreamId};
use isupipe_core::models::user::UserId;

#[derive(Debug, toasty::Model)]
#[table = "livestreams"]
pub struct LivestreamRow {
    #[key]
    #[auto]
    pub id: i64,
    pub user_id: i64,
    pub title: String,
    pub description: String,
    pub playlist_url: String,
    pub thumbnail_url: String,
    pub start_at: i64,
    pub end_at: i64,
}

impl From<LivestreamRow> for Livestream {
    fn from(row: LivestreamRow) -> Self {
        Self {
            id: LivestreamId::new(row.id),
            user_id: UserId::new(row.user_id),
            title: row.title,
            description: row.description,
            playlist_url: row.playlist_url,
            thumbnail_url: row.thumbnail_url,
            start_at: row.start_at,
            end_at: row.end_at,
        }
    }
}
