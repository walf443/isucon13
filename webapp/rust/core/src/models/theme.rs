use crate::models::user::UserId;
use fake::Dummy;
use kubetsu::Id;

pub const TABLE_NAME: &str = "themes";

#[derive(Debug, sqlx::FromRow, Dummy)]
pub struct Theme {
    pub id: Id<Self, i64>,
    #[allow(unused)]
    pub user_id: UserId,
    pub dark_mode: bool,
}

pub type ThemeId = Id<Theme, i64>;
