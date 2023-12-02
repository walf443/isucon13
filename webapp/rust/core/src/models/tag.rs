#[derive(Debug, serde::Serialize)]
pub struct Tag {
    pub id: i64,
    pub name: String,
}

#[derive(Debug, sqlx::FromRow)]
pub struct TagModel {
    pub id: i64,
    pub name: String,
}
