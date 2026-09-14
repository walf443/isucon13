use crate::models::livestream::LivestreamId;
use crate::models::user::UserId;
use fake::Dummy;

#[derive(Debug)]
pub struct Reaction {
    pub id: ReactionId,
    pub emoji_name: String,
    pub user_id: UserId,
    pub livestream_id: LivestreamId,
    pub created_at: i64,
}

#[derive(Dummy)]
pub struct CreateReaction {
    pub emoji_name: String,
    pub user_id: UserId,
    pub livestream_id: LivestreamId,
    pub created_at: i64,
}

kubetsu::define_id!(
    #[derive(toasty::Embed)]
    pub struct ReactionId(i64);
);
kubetsu_serde::impl_serde!(ReactionId(i64));
kubetsu_fake::impl_fake!(ReactionId(i64));
