use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::livestream_tag::{LivestreamTag, LivestreamTagId};
use isupipe_core::models::tag::TagId;

#[derive(Debug, toasty::Model)]
#[table = "livestream_tags"]
pub struct LivestreamTagRow {
    #[key]
    #[auto]
    pub id: LivestreamTagId,
    pub livestream_id: LivestreamId,
    pub tag_id: TagId,
}

impl From<LivestreamTagRow> for LivestreamTag {
    fn from(row: LivestreamTagRow) -> Self {
        Self {
            id: row.id,
            livestream_id: row.livestream_id,
            tag_id: row.tag_id,
        }
    }
}
