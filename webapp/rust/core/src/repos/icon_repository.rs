use crate::db::DBConn;
use crate::repos::Result;
use async_trait::async_trait;

#[async_trait]
pub trait IconRepository {
    async fn find_image_by_user_id(
        &self,
        conn: &mut DBConn,
        user_id: i64,
    ) -> Result<Option<Vec<u8>>>;

    async fn insert(&self, conn: &mut DBConn, user_id: i64, image: &Vec<u8>) -> Result<i64>;

    async fn delete_by_user_id(&self, conn: &mut DBConn, user_id: i64) -> Result<()>;
}
