use crate::models::livestream::LivestreamId;
use crate::models::user::UserId;

pub const TABLE_NAME: &str = "livestream_viewers_history";

pub struct CreateLivestreamViewersHistory {
    pub user_id: UserId,
    pub livestream_id: LivestreamId,
    pub created_at: i64,
}
