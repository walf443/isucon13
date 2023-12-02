#[derive(Debug, serde::Serialize)]
pub struct LivestreamStatistics {
    pub rank: i64,
    pub viewers_count: i64,
    pub total_reactions: i64,
    pub total_reports: i64,
    pub max_tip: i64,
}
