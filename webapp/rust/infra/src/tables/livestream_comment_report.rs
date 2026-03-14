use sqipe::{col, table, Col, TableRef};

pub struct LivestreamCommentReportTable;

pub const TABLE_LIVECOMMENT_REPORTS: LivestreamCommentReportTable = LivestreamCommentReportTable;

impl LivestreamCommentReportTable {
    pub fn table_name(&self) -> &'static str {
        "livecomment_reports"
    }

    pub fn table(&self) -> TableRef {
        table(self.table_name())
    }

    pub fn livestream_id(&self) -> Col {
        col("livestream_id")
    }
}
