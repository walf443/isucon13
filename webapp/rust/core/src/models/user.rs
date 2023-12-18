#[derive(Debug, sqlx::FromRow)]
pub struct User {
    pub id: i64,
    pub name: String,
    pub display_name: Option<String>,
    pub description: Option<String>,
    #[sqlx(default, rename = "password")]
    pub hashed_password: Option<String>,
}
