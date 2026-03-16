qbey_schema!(ReactionTable, "reactions", [
    id,
    user_id,
    livestream_id,
    emoji_name,
    created_at,
]);

pub const TABLE_REACTIONS: ReactionTable = ReactionTable::new();
