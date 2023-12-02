#[derive(Debug, serde::Serialize)]
pub struct UserStatistics {
    pub rank: i64,
    pub viewers_count: i64,
    pub total_reactions: i64,
    pub total_livecomments: i64,
    pub total_tip: i64,
    pub favorite_emoji: String,
}
