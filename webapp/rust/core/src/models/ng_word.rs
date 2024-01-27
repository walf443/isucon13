use crate::models::livestream::LivestreamId;
use crate::models::user::UserId;
use kubetsu::Id;

#[derive(Debug, serde::Serialize, sqlx::FromRow)]
pub struct NgWord {
    pub id: Id<Self, i64>,
    pub user_id: UserId,
    pub livestream_id: LivestreamId,
    pub word: String,
    #[sqlx(default)]
    pub created_at: i64,
}

pub type NgWordId = Id<NgWord, i64>;

pub struct CreateNgWord {
    pub user_id: UserId,
    pub livestream_id: LivestreamId,
    pub word: String,
    pub created_at: i64,
}
