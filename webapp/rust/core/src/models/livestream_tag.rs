use crate::models::id::Id;
use crate::models::livestream::LivestreamId;
use crate::models::tag::TagId;

#[derive(Debug, sqlx::FromRow)]
pub struct LivestreamTag {
    #[allow(unused)]
    pub id: Id<Self>,
    pub livestream_id: LivestreamId,
    pub tag_id: TagId,
}

pub type LivestreamTagId = Id<LivestreamTag>;
