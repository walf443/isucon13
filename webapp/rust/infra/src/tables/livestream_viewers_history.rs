use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::user::UserId;

#[derive(Debug, toasty::Model)]
#[table = "livestream_viewers_history"]
pub struct LivestreamViewersHistoryRow {
    #[key]
    #[auto]
    pub id: i64,
    pub user_id: UserId,
    pub livestream_id: LivestreamId,
    pub created_at: i64,
}
