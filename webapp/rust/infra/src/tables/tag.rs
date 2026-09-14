use isupipe_core::models::tag::{Tag, TagId, TagName};

#[derive(Debug, toasty::Model)]
#[table = "tags"]
pub struct TagRow {
    #[key]
    #[auto]
    pub id: TagId,
    #[unique]
    pub name: TagName,
}

impl From<TagRow> for Tag {
    fn from(row: TagRow) -> Self {
        Self {
            id: row.id,
            name: row.name,
        }
    }
}
