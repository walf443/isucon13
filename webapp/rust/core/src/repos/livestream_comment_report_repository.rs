use crate::db::DBConn;
use crate::models::livestream_comment_report::{
    LivestreamCommentReport, LivestreamCommentReportId,
};
use crate::repos::Result;
use async_trait::async_trait;

#[async_trait]
pub trait LivestreamCommentReportRepository {
    async fn insert(
        &self,
        conn: &mut DBConn,
        user_id: i64,
        livestream_id: i64,
        livestream_comment_id: i64,
        created_at: i64,
    ) -> Result<LivestreamCommentReportId>;
    async fn count_by_livestream_id(&self, conn: &mut DBConn, livestream_id: i64) -> Result<i64>;

    async fn find_all_by_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: i64,
    ) -> Result<Vec<LivestreamCommentReport>>;
}

pub trait HaveLivestreamCommentReportRepository {
    type Repo: Sync + LivestreamCommentReportRepository;
    fn livestream_comment_report_repo(&self) -> &Self::Repo;
}
