use crate::models::id::Id;
use crate::models::user::UserId;
use fake::Dummy;

#[derive(Debug, sqlx::FromRow, Dummy)]
pub struct Theme {
    pub id: Id<Self>,
    #[allow(unused)]
    pub user_id: UserId,
    pub dark_mode: bool,
}

pub type ThemeId = Id<Theme>;
