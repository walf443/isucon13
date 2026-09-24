use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::reaction::ReactionId;
use isupipe_core::models::user::UserId;

qbey::qbey_schema!(
    ReactionTable,
    "reactions",
    [
        id: ReactionId,
        user_id: UserId,
        livestream_id: LivestreamId,
        emoji_name: String,
        created_at: i64,
    ], row = ReactionRow);

pub const TABLE_REACTIONS: ReactionTable = ReactionTable::new();
