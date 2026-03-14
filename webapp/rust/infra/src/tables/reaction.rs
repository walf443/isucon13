sqipe_schema!(ReactionTable, "reactions", [
    id,
    user_id,
    livestream_id,
    created_at,
]);

pub const TABLE_REACTIONS: ReactionTable = ReactionTable::new();
