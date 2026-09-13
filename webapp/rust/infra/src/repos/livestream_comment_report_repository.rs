#[cfg(test)]
mod count_by_livestream_id;
#[cfg(test)]
mod create;
#[cfg(test)]
mod find_all_by_livestream_id;

use crate::sql_support::scalar_i64;
use crate::tables::livestream_comment_report::LivestreamCommentReportRow;
use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::livestream_comment_report::{
    CreateLivestreamCommentReport, LivestreamCommentReport, LivestreamCommentReportId,
};
use isupipe_core::repos::Result;
use isupipe_core::repos::livestream_comment_report_repository::LivestreamCommentReportRepository;

#[derive(Clone)]
pub struct LivestreamCommentReportRepositoryInfra {}

#[async_trait]
impl LivestreamCommentReportRepository for LivestreamCommentReportRepositoryInfra {
    async fn create<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        report: &CreateLivestreamCommentReport,
    ) -> Result<LivestreamCommentReportId> {
        let row = LivestreamCommentReportRow::create()
            .user_id(&report.user_id)
            .livestream_id(&report.livestream_id)
            .livecomment_id(&report.livestream_comment_id)
            .created_at(report.created_at)
            .exec(conn)
            .await?;

        Ok(row.id)
    }

    async fn count_by_livestream_id<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        livestream_id: &LivestreamId,
    ) -> isupipe_core::repos::Result<i64> {
        let rows = toasty::sql::query(
            r#"
            SELECT COUNT(*)
            FROM livestreams l
            INNER JOIN livecomment_reports r ON r.livestream_id = l.id
            WHERE l.id = ?
            "#,
        )
        .bind(*livestream_id.inner())
        .exec(conn)
        .await?;

        Ok(scalar_i64(rows)?)
    }

    async fn find_all_by_livestream_id<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        livestream_id: &LivestreamId,
    ) -> isupipe_core::repos::Result<Vec<LivestreamCommentReport>> {
        let rows = LivestreamCommentReportRow::filter(
            LivestreamCommentReportRow::fields()
                .livestream_id()
                .eq(livestream_id),
        )
        .exec(conn)
        .await?;

        Ok(rows.into_iter().map(Into::into).collect())
    }
}
