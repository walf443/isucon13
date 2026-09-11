qbey::qbey_schema!(
    LivestreamViewersHistoryTable,
    "livestream_viewers_history",
    [user_id, livestream_id,]
);

pub const TABLE_LIVESTREAM_VIEWERS_HISTORY: LivestreamViewersHistoryTable =
    LivestreamViewersHistoryTable::new();
