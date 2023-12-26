use crate::responses::livestream_comment_response::LivestreamCommentResponse;
use crate::responses::user_response::UserResponse;
use isupipe_core::db::DBConn;
use isupipe_core::models::livestream_comment_report::LivestreamCommentReport;
use isupipe_core::repos::livestream_comment_repository::LivestreamCommentRepository;
use isupipe_core::repos::user_repository::UserRepository;
use isupipe_http_core::responses::ResponseResult;
use isupipe_infra::repos::livestream_comment_repository::LivestreamCommentRepositoryInfra;
use isupipe_infra::repos::user_repository::UserRepositoryInfra;

#[derive(Debug, serde::Serialize)]
pub struct LivestreamCommentReportResponse {
    pub id: i64,
    pub reporter: UserResponse,
    pub livecomment: LivestreamCommentResponse,
    pub created_at: i64,
}

impl LivestreamCommentReportResponse {
    pub async fn build(
        conn: &mut DBConn,
        report_model: LivestreamCommentReport,
    ) -> ResponseResult<Self> {
        let user_repo = UserRepositoryInfra {};
        let reporter_model = user_repo
            .find(conn, report_model.user_id.get())
            .await?
            .unwrap();
        let reporter = UserResponse::build(conn, reporter_model).await?;

        let comment_repo = LivestreamCommentRepositoryInfra {};

        let comment_model = comment_repo
            .find(conn, &report_model.livestream_comment_id)
            .await?
            .unwrap();

        let livecomment = LivestreamCommentResponse::build(conn, comment_model).await?;

        Ok(Self {
            id: report_model.id.get(),
            reporter,
            livecomment,
            created_at: report_model.created_at,
        })
    }
}
