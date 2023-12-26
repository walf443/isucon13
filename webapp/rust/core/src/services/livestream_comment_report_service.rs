use crate::db::HaveDBPool;
use crate::models::livestream::LivestreamId;
use crate::models::livestream_comment::LivestreamCommentId;
use crate::models::livestream_comment_report::LivestreamCommentReport;
use crate::repos::livestream_comment_report_repository::{
    HaveLivestreamCommentReportRepository, LivestreamCommentReportRepository,
};
use crate::repos::livestream_comment_repository::{
    HaveLivestreamCommentRepository, LivestreamCommentRepository,
};
use crate::repos::livestream_repository::{HaveLivestreamRepository, LivestreamRepository};
use crate::services::ServiceError::{NotFoundLivestream, NotFoundLivestreamComment};
use crate::services::ServiceResult;
use async_trait::async_trait;
use chrono::Utc;
use crate::models::user::UserId;

#[async_trait]
pub trait LivestreamCommentReportService {
    async fn create(
        &self,
        user_id: &UserId,
        livestream_id: &LivestreamId,
        livestream_comment_id: &LivestreamCommentId,
    ) -> ServiceResult<LivestreamCommentReport>;
}

pub trait HaveLivestreamCommentReportService {
    type Service: LivestreamCommentReportService;

    fn livestream_comment_report_service(&self) -> &Self::Service;
}

#[async_trait]
pub trait LivestreamCommentReportServiceImpl:
    Sync
    + HaveDBPool
    + HaveLivestreamCommentReportRepository
    + HaveLivestreamRepository
    + HaveLivestreamCommentRepository
{
}

#[async_trait]
impl<S: LivestreamCommentReportServiceImpl> LivestreamCommentReportService for S {
    async fn create(
        &self,
        user_id: &UserId,
        livestream_id: &LivestreamId,
        livestream_comment_id: &LivestreamCommentId,
    ) -> ServiceResult<LivestreamCommentReport> {
        let pool = self.get_db_pool();
        let mut tx = pool.begin().await?;

        let livestream_repo = self.livestream_repo();
        livestream_repo
            .find(&mut *tx, livestream_id.get())
            .await?
            .ok_or(NotFoundLivestream)?;

        let comment_repo = self.livestream_comment_repo();
        let _ = comment_repo
            .find(&mut *tx, livestream_comment_id.get())
            .await?
            .ok_or(NotFoundLivestreamComment)?;

        let now = Utc::now().timestamp();
        let report_id = self
            .livestream_comment_report_repo()
            .insert(
                &mut *tx,
                user_id.get(),
                livestream_id.get(),
                livestream_comment_id,
                now,
            )
            .await?;

        tx.commit().await?;

        Ok(LivestreamCommentReport {
            id: report_id,
            user_id: user_id.get(),
            livestream_id: livestream_id.clone(),
            livestream_comment_id: livestream_comment_id.clone(),
            created_at: now,
        })
    }
}
