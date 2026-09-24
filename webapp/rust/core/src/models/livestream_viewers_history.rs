use crate::models::livestream::LivestreamId;
use crate::models::user::UserId;

#[derive(fake::Dummy)]
pub struct CreateLivestreamViewersHistory {
    pub user_id: UserId,
    pub livestream_id: LivestreamId,
    pub created_at: i64,
}
