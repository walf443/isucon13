use crate::db::DBConn;
use crate::repos::Result;
use async_trait::async_trait;

#[async_trait]
pub trait LivestreamViewersHistoryRepository {
    async fn delete_by_livestream_id_and_user_id(
        &self,
        conn: &mut DBConn,
        livestream_id: i64,
        user_id: i64,
    ) -> Result<()>;
}
