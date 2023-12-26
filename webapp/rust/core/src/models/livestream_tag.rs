use crate::models::id::Id;

#[derive(Debug, sqlx::FromRow)]
pub struct LivestreamTag {
    #[allow(unused)]
    pub id: Id<Self>,
    pub livestream_id: i64,
    pub tag_id: i64,
}

pub type LivestreamTagId = Id<LivestreamTag>;
