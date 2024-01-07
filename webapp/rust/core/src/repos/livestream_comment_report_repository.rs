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
    async fn create(
        &self,
        conn: &mut DBConn,
        report: &CreateLivestreamCommentReport,
    ) -> Result<LivestreamCommentReportId>;
    async fn count_by_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
    ) -> Result<i64>;

    async fn find_all_by_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
    ) -> Result<Vec<LivestreamCommentReport>>;
}

pub trait HaveLivestreamCommentReportRepository {
    type Repo: Sync + LivestreamCommentReportRepository;
    fn livestream_comment_report_repo(&self) -> &Self::Repo;
}
