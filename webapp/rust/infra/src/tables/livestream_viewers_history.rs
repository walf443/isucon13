use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::user::UserId;

qbey::qbey_schema!(
    LivestreamViewersHistoryTable,
    "livestream_viewers_history",
    [id: i64, user_id: UserId, livestream_id: LivestreamId, created_at: i64], row = LivestreamViewersHistoryRow);

pub const TABLE_LIVESTREAM_VIEWERS_HISTORY: LivestreamViewersHistoryTable =
    LivestreamViewersHistoryTable::new();
