#[derive(Debug, toasty::Model)]
#[table = "livestream_viewers_history"]
pub struct LivestreamViewersHistoryRow {
    #[key]
    #[auto]
    pub id: i64,
    pub user_id: i64,
    pub livestream_id: i64,
    pub created_at: i64,
}
