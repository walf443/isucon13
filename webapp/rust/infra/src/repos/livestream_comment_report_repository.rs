use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::livestream_comment_report::{
    CreateLivestreamCommentReport, LivestreamCommentReport, LivestreamCommentReportId,
};
use isupipe_core::models::mysql_decimal::MysqlDecimal;
use isupipe_core::repos::livestream_comment_report_repository::LivestreamCommentReportRepository;
use isupipe_core::repos::Result;

#[derive(Clone)]
pub struct LivestreamCommentReportRepositoryInfra {}

#[async_trait]
impl LivestreamCommentReportRepository for LivestreamCommentReportRepositoryInfra {
    async fn create(
        &self,
        conn: &mut DBConn,
        report: &CreateLivestreamCommentReport,
    ) -> Result<LivestreamCommentReportId> {
        let rs = sqlx::query(
            "INSERT INTO livecomment_reports(user_id, livestream_id, livecomment_id, created_at) VALUES (?, ?, ?, ?)",
        )
            .bind(&report.user_id)
            .bind(&report.livestream_id)
            .bind(&report.livestream_comment_id)
            .bind(report.created_at)
            .execute(conn)
            .await?;
        let report_id = rs.last_insert_id() as i64;
        Ok(LivestreamCommentReportId::new(report_id))
    }

    async fn count_by_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
    ) -> isupipe_core::repos::Result<i64> {
        let MysqlDecimal(total_reports) = sqlx::query_scalar("SELECT COUNT(*) FROM livestreams l INNER JOIN livecomment_reports r ON r.livestream_id = l.id WHERE l.id = ?")
            .bind(livestream_id)
            .fetch_one(conn)
            .await?;

        Ok(total_reports)
    }

    async fn find_all_by_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
    ) -> isupipe_core::repos::Result<Vec<LivestreamCommentReport>> {
        let report_models: Vec<LivestreamCommentReport> =
            sqlx::query_as("SELECT * FROM livecomment_reports WHERE livestream_id = ?")
                .bind(livestream_id)
                .fetch_all(conn)
                .await?;

        Ok(report_models)
    }
}
