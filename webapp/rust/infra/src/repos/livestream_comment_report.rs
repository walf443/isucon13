use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::livestream_comment_report::LivecommentReportModel;
use isupipe_core::repos::livestream_comment_report::LivestreamCommentReportRepository;

pub struct LivestreamCommentReportRepositoryInfra {}

#[async_trait]
impl LivestreamCommentReportRepository for LivestreamCommentReportRepositoryInfra {
    async fn find_all_by_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: i64,
    ) -> isupipe_core::repos::Result<Vec<LivecommentReportModel>> {
        let report_models: Vec<LivecommentReportModel> =
            sqlx::query_as("SELECT * FROM livecomment_reports WHERE livestream_id = ?")
                .bind(livestream_id)
                .fetch_all(conn)
                .await?;

        Ok(report_models)
    }
}
