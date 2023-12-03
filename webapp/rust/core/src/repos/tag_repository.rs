use crate::db::DBConn;
use crate::models::tag::TagModel;
use crate::repos::Result;
use async_trait::async_trait;

#[async_trait]
pub trait TagRepository {
    async fn find_all(&self, conn: &mut DBConn) -> Result<Vec<TagModel>>;
}
