use crate::models::user::UserId;

#[derive(fake::Dummy)]
pub struct CreateIcon {
    pub user_id: UserId,
    pub image: Vec<u8>,
}

kubetsu::define_id!(
    #[derive(toasty::Embed)]
    pub struct IconId(i64);
);
kubetsu_serde::impl_serde!(IconId(i64));
kubetsu_fake::impl_fake!(IconId(i64));
