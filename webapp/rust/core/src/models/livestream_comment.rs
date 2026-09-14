use crate::models::livestream::LivestreamId;
use crate::models::user::UserId;

#[derive(Debug)]
pub struct LivestreamComment {
    pub id: LivestreamCommentId,
    pub user_id: UserId,
    pub livestream_id: LivestreamId,
    pub comment: String,
    pub tip: i64,
    pub created_at: i64,
}

#[derive(fake::Dummy)]
pub struct CreateLivestreamComment {
    pub user_id: UserId,
    pub livestream_id: LivestreamId,
    pub comment: String,
    pub tip: i64,
    pub created_at: i64,
}

kubetsu::define_id!(
    #[derive(toasty::Embed)]
    pub struct LivestreamCommentId(i64);
);
kubetsu_serde::impl_serde!(LivestreamCommentId(i64));
kubetsu_fake::impl_fake!(LivestreamCommentId(i64));
