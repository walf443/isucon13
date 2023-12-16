use crate::db::DBConn;
use crate::repos::Result;
use async_trait::async_trait;

#[async_trait]
pub trait LivestreamCommentRepository {
    async fn insert(
        &self,
        conn: &mut DBConn,
        user_id: i64,
        livestream_id: i64,
        comment: &str,
        tip: i64,
        created_at: i64,
    ) -> Result<i64>;
}
