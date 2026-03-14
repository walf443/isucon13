use sqipe::{table, Col, TableRef};

pub struct LivestreamTagTable;

pub const TABLE_LIVESTREAM_TAGS: LivestreamTagTable = LivestreamTagTable;

impl LivestreamTagTable {
    pub fn table_name(&self) -> &'static str {
        "livestream_tags"
    }

    pub fn table(&self) -> TableRef {
        table(self.table_name())
    }

    pub fn livestream_id(&self) -> Col {
        self.table().col("livestream_id")
    }

    pub fn tag_id(&self) -> Col {
        self.table().col("tag_id")
    }
}
