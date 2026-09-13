use crate::models::livestream::LivestreamId;
use crate::models::livestream_comment::LivestreamCommentId;
use crate::models::user::UserId;

#[derive(Debug)]
pub struct LivestreamCommentReport {
    pub id: LivestreamCommentReportId,
    pub user_id: UserId,
    pub livestream_id: LivestreamId,
    pub livestream_comment_id: LivestreamCommentId,
    pub created_at: i64,
}

#[derive(fake::Dummy)]
pub struct CreateLivestreamCommentReport {
    pub user_id: UserId,
    pub livestream_id: LivestreamId,
    pub livestream_comment_id: LivestreamCommentId,
    pub created_at: i64,
}

kubetsu::define_id!(
    #[derive(toasty::Embed)]
    pub struct LivestreamCommentReportId(i64);
);
kubetsu_serde::impl_serde!(LivestreamCommentReportId(i64));
kubetsu_fake::impl_fake!(LivestreamCommentReportId(i64));
