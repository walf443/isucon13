sqipe_schema!(LivestreamCommentTable, TABLE_LIVECOMMENTS, "livecomments", [
    id,
    user_id,
    livestream_id,
    created_at,
]);
