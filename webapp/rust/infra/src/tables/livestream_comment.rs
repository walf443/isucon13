use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::livestream_comment::LivestreamCommentId;
use isupipe_core::models::user::UserId;

qbey::qbey_schema!(
    LivestreamCommentTable,
    "livecomments",
    [
        id: LivestreamCommentId,
        user_id: UserId,
        livestream_id: LivestreamId,
        comment: String,
        tip: i64,
        created_at: i64,
    ], row = LivestreamCommentRow);

pub const TABLE_LIVECOMMENTS: LivestreamCommentTable = LivestreamCommentTable::new();
