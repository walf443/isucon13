use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::ng_word::NgWordId;
use isupipe_core::models::user::UserId;

qbey::qbey_schema!(
    NgWordTable,
    "ng_words",
    [
        id: NgWordId,
        user_id: UserId,
        livestream_id: LivestreamId,
        word: String,
        created_at: i64,
    ]
);

pub const TABLE_NG_WORDS: NgWordTable = NgWordTable::new();
