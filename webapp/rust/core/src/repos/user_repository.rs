use crate::db::DBConn;
use crate::models::user::UserModel;
use crate::repos::Result;
use async_trait::async_trait;

#[async_trait]
pub trait UserRepository {
    async fn find_id_by_name(&self, conn: &mut DBConn, name: &str) -> Result<Option<i64>>;
    async fn find_by_name(&self, conn: &mut DBConn, name: &str) -> Result<Option<UserModel>>;
}
