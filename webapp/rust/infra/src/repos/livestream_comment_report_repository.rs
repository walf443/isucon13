#[cfg(test)]
mod count_by_livestream_id;
#[cfg(test)]
mod find_all_by_livestream_id;

use crate::qbey_support::bind_qbey_values;
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
use qbey_mysql::qbey;

#[derive(Clone)]
pub struct LivestreamCommentReportRepositoryInfra {}

#[async_trait]
impl LivestreamCommentReportRepository for LivestreamCommentReportRepositoryInfra {
    async fn create(
        &self,
        conn: &mut DBConn,
        report: &CreateLivestreamCommentReport,
    ) -> Result<LivestreamCommentReportId> {
        let t = &TABLE_LIVECOMMENT_REPORTS;
        let mut ins = qbey(t.table()).into_insert();
        ins.add_value(&[
            ("user_id", (*report.user_id.inner()).into()),
            ("livestream_id", (*report.livestream_id.inner()).into()),
            ("livecomment_id", (*report.livestream_comment_id.inner()).into()),
            ("created_at", report.created_at.into()),
        ]);
        let (sql, binds) = ins.to_sql();
        let rs = bind_qbey_values!(sqlx::query(&sql), binds)
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
        let mut q = qbey(livestream.table());
        q.as_("l");
        let l = livestream.as_("l");
        q.join(
            report.table(),
            report.livestream_id().eq_col(l.id()),
        );
        q.and_where(l.id().eq(*livestream_id.inner()));
        q.add_select(qbey::count_all());
        let (sql, binds) = q.to_sql();
        let total_reports = bind_qbey_values!(sqlx::query_scalar::<_, i64>(&sql), binds)
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
        let mut q = qbey(t.table());
        q.and_where(t.livestream_id().eq(*livestream_id.inner()));
        let (sql, binds) = q.to_sql();
        let report_models =
            bind_qbey_values!(sqlx::query_as::<_, LivestreamCommentReport>(&sql), binds)
                .fetch_all(conn)
                .await?;

        Ok(report_models)
    }
}
