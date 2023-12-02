use crate::models::theme::Theme;

#[derive(Debug, sqlx::FromRow)]
pub struct UserModel {
    pub id: i64,
    pub name: String,
    pub display_name: Option<String>,
    pub description: Option<String>,
    #[sqlx(default, rename = "password")]
    pub hashed_password: Option<String>,
}

#[derive(Debug, serde::Serialize)]
pub struct User {
    pub id: i64,
    pub name: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub display_name: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub description: Option<String>,
    pub theme: Theme,
    pub icon_hash: String,
}
