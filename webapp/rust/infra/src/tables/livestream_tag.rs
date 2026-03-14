sqipe_schema!(LivestreamTagTable, "livestream_tags", [
    livestream_id,
    tag_id,
]);

pub const TABLE_LIVESTREAM_TAGS: LivestreamTagTable = LivestreamTagTable::new();
