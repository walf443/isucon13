use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::ng_word::{NgWord, NgWordId};
use isupipe_core::models::user::UserId;

#[derive(Debug, toasty::Model)]
#[table = "ng_words"]
pub struct NgWordRow {
    #[key]
    #[auto]
    pub id: i64,
    pub user_id: i64,
    pub livestream_id: i64,
    #[index]
    pub word: String,
    pub created_at: i64,
}

impl From<NgWordRow> for NgWord {
    fn from(row: NgWordRow) -> Self {
        Self {
            id: NgWordId::new(row.id),
            user_id: UserId::new(row.user_id),
            livestream_id: LivestreamId::new(row.livestream_id),
            word: row.word,
            created_at: row.created_at,
        }
    }
}
