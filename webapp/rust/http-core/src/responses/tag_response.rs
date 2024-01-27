use isupipe_core::models::tag::{Tag, TagId, TagName};

#[derive(Debug, serde::Serialize)]
pub struct TagResponse {
    pub id: TagId,
    pub name: TagName,
}

impl From<Tag> for TagResponse {
    fn from(tag: Tag) -> Self {
        Self {
            id: tag.id.clone(),
            name: tag.name.clone(),
        }
    }
}
