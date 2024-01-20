use isupipe_core::models::tag::Tag;

#[derive(Debug, serde::Serialize)]
pub struct TagResponse {
    pub id: i64,
    pub name: String,
}

impl From<Tag> for TagResponse {
    fn from(tag: Tag) -> Self {
        Self {
            id: tag.id.get(),
            name: tag.name.get(),
        }
    }
}
