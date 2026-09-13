use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::livestream_tag::{LivestreamTag, LivestreamTagId};
use isupipe_core::models::tag::TagId;

#[derive(Debug, toasty::Model)]
#[table = "livestream_tags"]
pub struct LivestreamTagRow {
    #[key]
    #[auto]
    pub id: i64,
    pub livestream_id: i64,
    pub tag_id: i64,
}

impl From<LivestreamTagRow> for LivestreamTag {
    fn from(row: LivestreamTagRow) -> Self {
        Self {
            id: LivestreamTagId::new(row.id),
            livestream_id: LivestreamId::new(row.livestream_id),
            tag_id: TagId::new(row.tag_id),
        }
    }
}
