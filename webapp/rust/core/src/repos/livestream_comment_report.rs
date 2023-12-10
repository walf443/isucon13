use crate::db::DBConn;
use crate::models::livestream_comment_report::LivecommentReportModel;
use crate::repos::Result;
use async_trait::async_trait;

#[async_trait]
pub trait LivestreamCommentReportRepository {
    async fn find_all_by_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: i64,
    ) -> Result<Vec<LivecommentReportModel>>;
}
