
#[derive(Debug, sqlx::FromRow)]
pub struct TagModel {
    pub id: i64,
    pub name: String,
}