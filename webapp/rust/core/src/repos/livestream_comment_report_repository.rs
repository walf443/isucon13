use crate::db::DBConn;
use crate::models::livestream::LivestreamId;
use crate::models::livestream_comment_report::{
    CreateLivestreamCommentReport, LivestreamCommentReport, LivestreamCommentReportId,
};
use crate::repos::Result;
use async_trait::async_trait;

#[cfg_attr(any(feature = "test", test), mockall::automock)]
#[async_trait]
pub trait LivestreamCommentReportRepository {
    async fn create<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        report: &CreateLivestreamCommentReport,
    ) -> Result<LivestreamCommentReportId>;
    async fn count_by_livestream_id<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        livestream_id: &LivestreamId,
    ) -> Result<i64>;

    async fn find_all_by_livestream_id<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        livestream_id: &LivestreamId,
    ) -> Result<Vec<LivestreamCommentReport>>;
}

pub trait HaveLivestreamCommentReportRepository {
    type Repo: Sync + LivestreamCommentReportRepository;
    fn livestream_comment_report_repo(&self) -> &Self::Repo;
}
