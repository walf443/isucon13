qbey::qbey_schema!(NgWordTable, "ng_words", [
    id,
    user_id,
    livestream_id,
    word,
    created_at,
]);

pub const TABLE_NG_WORDS: NgWordTable = NgWordTable::new();
