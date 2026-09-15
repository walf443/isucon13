use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::livestream_tag::LivestreamTagId;
use isupipe_core::models::tag::TagId;

qbey::qbey_schema!(
    LivestreamTagTable,
    "livestream_tags",
    [id: LivestreamTagId, livestream_id: LivestreamId, tag_id: TagId], row = LivestreamTagRow);

pub const TABLE_LIVESTREAM_TAGS: LivestreamTagTable = LivestreamTagTable::new();
