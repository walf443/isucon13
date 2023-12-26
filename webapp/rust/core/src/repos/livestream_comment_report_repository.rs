use crate::db::DBConn;
use crate::models::livestream_comment::LivestreamCommentId;
use crate::models::livestream_comment_report::{
    LivestreamCommentReport, LivestreamCommentReportId,
};
use crate::repos::Result;
use async_trait::async_trait;
use crate::models::livestream::LivestreamId;
use crate::models::user::UserId;

#[async_trait]
pub trait LivestreamCommentReportRepository {
    async fn insert(
        &self,
        conn: &mut DBConn,
        user_id: &UserId,
        livestream_id: &LivestreamId,
        livestream_comment_id: &LivestreamCommentId,
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
