qbey::qbey_schema!(
    LivestreamCommentReportTable,
    "livecomment_reports",
    [id, livestream_id,]
);

pub const TABLE_LIVECOMMENT_REPORTS: LivestreamCommentReportTable =
    LivestreamCommentReportTable::new();
