use sqipe::{col, Col};

pub struct LivestreamViewersHistoryTable;

pub const TABLE_LIVESTREAM_VIEWERS_HISTORY: LivestreamViewersHistoryTable =
    LivestreamViewersHistoryTable;

impl LivestreamViewersHistoryTable {
    pub fn table_name(&self) -> &'static str {
        "livestream_viewers_history"
    }

    pub fn user_id(&self) -> Col {
        col("user_id")
    }

    pub fn livestream_id(&self) -> Col {
        col("livestream_id")
    }
}
