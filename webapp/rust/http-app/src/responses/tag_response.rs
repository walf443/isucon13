use isupipe_core::models::tag::TagModel;

#[derive(Debug, serde::Serialize)]
pub struct TagResponse {
    pub id: i64,
    pub name: String,
}

impl From<TagModel> for TagResponse {
    fn from(tag: TagModel) -> Self {
        Self {
            id: tag.id,
            name: tag.name,
        }
    }
}
