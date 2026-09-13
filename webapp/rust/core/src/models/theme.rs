use crate::models::user::UserId;
use fake::Dummy;

#[derive(Debug, Dummy)]
pub struct Theme {
    pub id: ThemeId,
    #[allow(unused)]
    pub user_id: UserId,
    pub dark_mode: bool,
}

kubetsu::define_id!(
    #[derive(toasty::Embed)]
    pub struct ThemeId(i64);
);
kubetsu_serde::impl_serde!(ThemeId(i64));
kubetsu_fake::impl_fake!(ThemeId(i64));
