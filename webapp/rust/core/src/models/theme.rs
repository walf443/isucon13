#[derive(Debug, serde::Deserialize, serde::Serialize)]
pub struct Theme {
    pub id: i64,
    pub dark_mode: bool,
}

#[derive(Debug, sqlx::FromRow)]
pub struct ThemeModel {
    pub id: i64,
    #[allow(unused)]
    pub user_id: i64,
    pub dark_mode: bool,
}