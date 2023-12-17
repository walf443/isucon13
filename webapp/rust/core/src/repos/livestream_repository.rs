use crate::db::DBConn;
use crate::models::livestream::{CreateLivestreamModel, LivestreamModel};
use crate::repos::Result;
use async_trait::async_trait;

#[async_trait]
pub trait LivestreamRepository {
    async fn create(&self, conn: &mut DBConn, stream: &CreateLivestreamModel) -> Result<i64>;
    async fn find_all(&self, conn: &mut DBConn) -> Result<Vec<LivestreamModel>>;

    async fn find_all_order_by_id_desc(&self, conn: &mut DBConn) -> Result<Vec<LivestreamModel>>;
    async fn find_all_order_by_id_desc_limit(
        &self,
        conn: &mut DBConn,
        limit: i64,
    ) -> Result<Vec<LivestreamModel>>;

    async fn find_all_by_user_id(
        &self,
        conn: &mut DBConn,
        user_id: i64,
    ) -> Result<Vec<LivestreamModel>>;
    async fn find(&self, conn: &mut DBConn, id: i64) -> Result<Option<LivestreamModel>>;

    async fn exist_by_id_and_user_id(
        &self,
        conn: &mut DBConn,
        id: i64,
        user_id: i64,
    ) -> Result<bool>;
}
