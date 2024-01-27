use crate::models::livestream::LivestreamId;
use crate::models::tag::TagId;
use kubetsu::Id;

#[derive(Debug, sqlx::FromRow)]
pub struct LivestreamTag {
    #[allow(unused)]
    pub id: Id<Self, i64>,
    pub livestream_id: LivestreamId,
    pub tag_id: TagId,
}

pub type LivestreamTagId = Id<LivestreamTag, i64>;
