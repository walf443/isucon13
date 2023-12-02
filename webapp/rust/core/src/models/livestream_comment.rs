use crate::models::livestream::Livestream;
use crate::models::user::User;

#[derive(Debug, serde::Serialize)]
pub struct Livecomment {
    pub id: i64,
    pub user: User,
    pub livestream: Livestream,
    pub comment: String,
    pub tip: i64,
    pub created_at: i64,
}

#[derive(Debug, sqlx::FromRow)]
pub struct LivecommentModel {
    pub id: i64,
    pub user_id: i64,
    pub livestream_id: i64,
    pub comment: String,
    pub tip: i64,
    pub created_at: i64,
}
