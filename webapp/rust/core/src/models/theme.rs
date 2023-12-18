#[derive(Debug, sqlx::FromRow)]
pub struct Theme {
    pub id: i64,
    #[allow(unused)]
    pub user_id: i64,
    pub dark_mode: bool,
}
