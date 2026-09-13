use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::reaction::{Reaction, ReactionId};
use isupipe_core::models::user::UserId;

#[derive(Debug, toasty::Model)]
#[table = "reactions"]
pub struct ReactionRow {
    #[key]
    #[auto]
    pub id: i64,
    pub user_id: i64,
    pub livestream_id: i64,
    pub emoji_name: String,
    pub created_at: i64,
}

impl From<ReactionRow> for Reaction {
    fn from(row: ReactionRow) -> Self {
        Self {
            id: ReactionId::new(row.id),
            emoji_name: row.emoji_name,
            user_id: UserId::new(row.user_id),
            livestream_id: LivestreamId::new(row.livestream_id),
            created_at: row.created_at,
        }
    }
}
