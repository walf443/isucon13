#[cfg(test)]
mod count_by_livestream_id;
#[cfg(test)]
mod find_all_by_livestream_id;

use crate::sqipe_support::bind_sqipe_values;
use crate::tables::livestream::TABLE_LIVESTREAMS;
use crate::tables::livestream_comment_report::TABLE_LIVECOMMENT_REPORTS;
use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::livestream_comment_report::{
    CreateLivestreamCommentReport, LivestreamCommentReport, LivestreamCommentReportId,
};
use isupipe_core::repos::livestream_comment_report_repository::LivestreamCommentReportRepository;
use isupipe_core::repos::Result;
use sqipe::{aggregate, table};
use sqipe_mysql::sqipe;

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
        let livestream = &TABLE_LIVESTREAMS;
        let report = &TABLE_LIVECOMMENT_REPORTS;
        let mut q = sqipe(livestream.table_name());
        q.as_("l");
        q.join(
            report.table_name(),
            report.table().col("livestream_id").eq_col(table("l").col("id")),
        );
        q.and_where(table("l").col("id").eq(*livestream_id.inner()));
        q.aggregate(&[aggregate::count_all()]);
        let (sql, binds) = q.to_sql();
        let total_reports = bind_sqipe_values!(sqlx::query_scalar::<_, i64>(&sql), binds)
            .fetch_one(conn)
            .await?;

        Ok(total_reports)
    }

    async fn find_all_by_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
    ) -> isupipe_core::repos::Result<Vec<LivestreamCommentReport>> {
        let t = &TABLE_LIVECOMMENT_REPORTS;
        let mut q = sqipe(t.table_name());
        q.and_where(t.livestream_id().eq(*livestream_id.inner()));
        let (sql, binds) = q.to_sql();
        let report_models =
            bind_sqipe_values!(sqlx::query_as::<_, LivestreamCommentReport>(&sql), binds)
                .fetch_all(conn)
                .await?;

        Ok(report_models)
    }
}
