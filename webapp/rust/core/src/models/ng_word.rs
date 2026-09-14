use crate::models::livestream::LivestreamId;
use crate::models::user::UserId;

#[derive(Debug, serde::Serialize)]
pub struct NgWord {
    pub id: NgWordId,
    pub user_id: UserId,
    pub livestream_id: LivestreamId,
    pub word: String,
    pub created_at: i64,
}

#[derive(fake::Dummy)]
pub struct CreateNgWord {
    pub user_id: UserId,
    pub livestream_id: LivestreamId,
    pub word: String,
    pub created_at: i64,
}

kubetsu::define_id!(
    #[derive(toasty::Embed)]
    pub struct NgWordId(i64);
);
kubetsu_serde::impl_serde!(NgWordId(i64));
kubetsu_fake::impl_fake!(NgWordId(i64));
