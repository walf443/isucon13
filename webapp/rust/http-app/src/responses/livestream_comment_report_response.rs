use crate::responses::livestream_comment_response::LivestreamCommentResponse;
use crate::responses::user_response::UserResponse;
use isupipe_core::models::livestream_comment::LivestreamComment;
use isupipe_core::models::livestream_comment_report::LivestreamCommentReport;
use isupipe_core::services::livestream_comment_service::LivestreamCommentService;
use isupipe_core::services::manager::ServiceManager;
use isupipe_core::services::user_service::UserService;
use isupipe_http_core::responses::ResponseResult;

#[derive(Debug, serde::Serialize)]
pub struct LivestreamCommentReportResponse {
    pub id: i64,
    pub reporter: UserResponse,
    pub livecomment: LivestreamCommentResponse,
    pub created_at: i64,
}

impl LivestreamCommentReportResponse {
    pub async fn build_by_service<S: ServiceManager>(
        service: &S,
        report_model: &LivestreamCommentReport,
    ) -> ResponseResult<Self> {
        let reporter_model = service
            .user_service()
            .find(&report_model.user_id)
            .await?
            .unwrap();
        let reporter = UserResponse::build_by_service(service, &reporter_model).await?;

        let comment_service = service.livestream_comment_service();
        let comment_model: LivestreamComment = comment_service
            .find(&report_model.livestream_comment_id)
            .await?
            .unwrap();

        let livecomment =
            LivestreamCommentResponse::build_by_service(service, &comment_model).await?;

        Ok(Self {
            id: report_model.id.get(),
            reporter,
            livecomment,
            created_at: report_model.created_at,
        })
    }
}
