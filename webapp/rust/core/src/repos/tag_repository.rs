use crate::db::DBConn;
use crate::models::tag::{Tag, TagId};
use crate::repos::Result;
use async_trait::async_trait;

#[async_trait]
pub trait TagRepository {
    async fn find(&self, conn: &mut DBConn, id: &TagId) -> Result<Tag>;
    async fn find_all(&self, conn: &mut DBConn) -> Result<Vec<Tag>>;

    async fn find_ids_by_name(&self, conn: &mut DBConn, name: &str) -> Result<Vec<TagId>>;
}
