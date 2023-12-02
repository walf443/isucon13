use crate::models::livestream_comment::Livecomment;
use crate::models::user::User;

#[derive(Debug, serde::Serialize)]
pub struct LivecommentReport {
    pub id: i64,
    pub reporter: User,
    pub livecomment: Livecomment,
    pub created_at: i64,
}

#[derive(Debug, sqlx::FromRow)]
pub struct LivecommentReportModel {
    pub id: i64,
    pub user_id: i64,
    #[allow(unused)]
    pub livestream_id: i64,
    pub livecomment_id: i64,
    pub created_at: i64,
}
