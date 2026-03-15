qbey_schema!(LivestreamCommentTable, "livecomments", [
    id,
    user_id,
    livestream_id,
    created_at,
]);

pub const TABLE_LIVECOMMENTS: LivestreamCommentTable = LivestreamCommentTable::new();
