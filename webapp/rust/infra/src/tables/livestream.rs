use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::user::UserId;

qbey::qbey_schema!(
    LivestreamTable,
    "livestreams",
    [
        id: LivestreamId,
        user_id: UserId,
        title: String,
        description: String,
        playlist_url: String,
        thumbnail_url: String,
        start_at: i64,
        end_at: i64,
    ], row = LivestreamRow);

pub const TABLE_LIVESTREAMS: LivestreamTable = LivestreamTable::new();
