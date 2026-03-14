use sqipe::{col, Col};

pub struct LivestreamCommentReportTable;

pub const TABLE_LIVECOMMENT_REPORTS: LivestreamCommentReportTable = LivestreamCommentReportTable;

impl LivestreamCommentReportTable {
    pub fn table_name(&self) -> &'static str {
        "livecomment_reports"
    }

    pub fn livestream_id(&self) -> Col {
        col("livestream_id")
    }
}
