use isupipe_core::models::tag::{Tag, TagId, TagName};

#[derive(Debug, toasty::Model)]
#[table = "tags"]
pub struct TagRow {
    #[key]
    #[auto]
    pub id: i64,
    #[unique]
    pub name: String,
}

impl From<TagRow> for Tag {
    fn from(row: TagRow) -> Self {
        Self {
            id: TagId::new(row.id),
            name: TagName::new(row.name),
        }
    }
}
