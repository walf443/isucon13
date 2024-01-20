use crate::models::id::Id;
use fake::Dummy;

#[derive(Debug, sqlx::FromRow, Dummy)]
pub struct User {
    pub id: Id<Self, i64>,
    pub name: String,
    pub display_name: Option<String>,
    pub description: Option<String>,
    #[sqlx(default, rename = "password")]
    pub hashed_password: Option<String>,
}

pub type UserId = Id<User, i64>;

#[derive(Debug, Dummy, PartialEq, Clone)]
pub struct CreateUser {
    pub name: String,
    pub display_name: String,
    pub description: String,
    pub password: String,
}
