use crate::db::DBConn;
use crate::models::livestream::LivestreamModel;
use crate::repos::Result;
use async_trait::async_trait;

#[async_trait]
pub trait LivestreamRepository {
    async fn find_all_by_user_id(
        &self,
        conn: &mut DBConn,
        user_id: i64,
    ) -> Result<Vec<LivestreamModel>>;
    async fn find(&self, conn: &mut DBConn, id: i64) -> Result<Option<LivestreamModel>>;
}
