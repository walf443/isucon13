use crate::models::id::Id;
use fake::Dummy;

#[derive(Debug, sqlx::FromRow, Dummy)]
pub struct Tag {
    pub id: Id<Self, i64>,
    pub name: TagName,
}

pub type TagId = Id<Tag, i64>;
pub type TagName = Id<Tag, String>;
