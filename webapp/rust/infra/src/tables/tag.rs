use isupipe_core::models::tag::{TagId, TagName};

qbey::qbey_schema!(TagTable, "tags", [id: TagId, name: TagName]);

pub const TABLE_TAGS: TagTable = TagTable::new();
