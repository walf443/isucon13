use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::reaction::{Reaction, ReactionId};
use isupipe_core::models::user::UserId;

#[derive(Debug, toasty::Model)]
#[table = "reactions"]
pub struct ReactionRow {
    #[key]
    #[auto]
    pub id: ReactionId,
    pub user_id: UserId,
    pub livestream_id: LivestreamId,
    pub emoji_name: String,
    pub created_at: i64,
}

impl From<ReactionRow> for Reaction {
    fn from(row: ReactionRow) -> Self {
        Self {
            id: row.id,
            emoji_name: row.emoji_name,
            user_id: row.user_id,
            livestream_id: row.livestream_id,
            created_at: row.created_at,
        }
    }
}
