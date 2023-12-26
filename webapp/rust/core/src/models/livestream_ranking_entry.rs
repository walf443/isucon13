use crate::models::livestream::LivestreamId;

#[derive(Debug)]
pub struct LivestreamRankingEntry {
    pub livestream_id: LivestreamId,
    pub score: i64,
}
