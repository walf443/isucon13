use crate::models::id::Id;
use crate::models::livestream::LivestreamId;
use crate::models::user::UserId;

#[derive(Debug, serde::Serialize, sqlx::FromRow)]
pub struct NgWord {
    pub id: Id<Self>,
    pub user_id: UserId,
    pub livestream_id: LivestreamId,
    pub word: String,
    #[sqlx(default)]
    pub created_at: i64,
}

pub type NgWordId = Id<NgWord>;

pub struct CreateNgWord {
    pub user_id: UserId,
    pub livestream_id: LivestreamId,
    pub word: String,
    pub created_at: i64,
}