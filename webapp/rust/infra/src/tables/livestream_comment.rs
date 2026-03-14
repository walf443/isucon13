use sqipe::{table, Col, TableRef};

pub struct LivestreamCommentTable;

pub const TABLE_LIVECOMMENTS: LivestreamCommentTable = LivestreamCommentTable;

impl LivestreamCommentTable {
    pub fn table_name(&self) -> &'static str {
        "livecomments"
    }

    pub fn table(&self) -> TableRef {
        table(self.table_name())
    }

    pub fn id(&self) -> Col {
        self.table().col("id")
    }

    pub fn user_id(&self) -> Col {
        self.table().col("user_id")
    }

    pub fn livestream_id(&self) -> Col {
        self.table().col("livestream_id")
    }

    pub fn created_at(&self) -> Col {
        self.table().col("created_at")
    }
}
