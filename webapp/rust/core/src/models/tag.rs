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

impl From<TagModel> for Tag {
    fn from(tag: TagModel) -> Self {
        Self {
            id: tag.id,
            name: tag.name,
        }
    }
}
