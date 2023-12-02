use crate::models::livestream::Livestream;
use crate::models::user::User;

#[derive(Debug, serde::Serialize)]
pub struct LivestreamComment {
    pub id: i64,
    pub user: User,
    pub livestream: Livestream,
    pub comment: String,
    pub tip: i64,
    pub created_at: i64,
}

#[derive(Debug, sqlx::FromRow)]
pub struct LivestreamCommentModel {
    pub id: i64,
    pub user_id: i64,
    pub livestream_id: i64,
    pub comment: String,
    pub tip: i64,
    pub created_at: i64,
}
