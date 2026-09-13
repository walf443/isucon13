use crate::models::user::UserId;
use fake::Dummy;

#[derive(Dummy)]
pub struct CreateLivestream {
    pub user_id: UserId,
    pub title: String,
    pub description: String,
    pub playlist_url: String,
    pub thumbnail_url: String,
    pub start_at: i64,
    pub end_at: i64,
}

#[derive(Debug)]
pub struct Livestream {
    pub id: LivestreamId,
    pub user_id: UserId,
    pub title: String,
    pub description: String,
    pub playlist_url: String,
    pub thumbnail_url: String,
    pub start_at: i64,
    pub end_at: i64,
}

kubetsu::define_id!(
    #[derive(toasty::Embed)]
    pub struct LivestreamId(i64);
);
kubetsu_serde::impl_serde!(LivestreamId(i64));
kubetsu_fake::impl_fake!(LivestreamId(i64));
