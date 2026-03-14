use sqipe::{table, Col, TableRef};

pub struct LivestreamViewersHistoryTable;

pub const TABLE_LIVESTREAM_VIEWERS_HISTORY: LivestreamViewersHistoryTable =
    LivestreamViewersHistoryTable;

impl LivestreamViewersHistoryTable {
    pub fn table_name(&self) -> &'static str {
        "livestream_viewers_history"
    }

    pub fn table(&self) -> TableRef {
        table(self.table_name())
    }

    pub fn user_id(&self) -> Col {
        self.table().col("user_id")
    }

    pub fn livestream_id(&self) -> Col {
        self.table().col("livestream_id")
    }
}
