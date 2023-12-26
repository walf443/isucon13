use crate::models::id::Id;

#[derive(Debug, sqlx::FromRow)]
pub struct Tag {
    pub id: Id<Self>,
    pub name: String,
}

pub type TagId = Id<Tag>;
