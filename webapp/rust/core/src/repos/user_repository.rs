use crate::db::DBConn;
use crate::models::user::UserModel;
use crate::repos::Result;
use async_trait::async_trait;

#[async_trait]
pub trait UserRepository {
    async fn insert(
        &self,
        conn: &mut DBConn,
        name: &str,
        display_name: &str,
        description: &str,
        password: &str,
    ) -> Result<i64>;

    fn hash_password(&self, password: &str) -> Result<String> {
        const BCRYPT_DEFAULT_COST: u32 = 4;
        let hashed_password = bcrypt::hash(password, BCRYPT_DEFAULT_COST)?;
        Ok(hashed_password)
    }

    async fn find(&self, conn: &mut DBConn, id: i64) -> Result<Option<UserModel>>;
    async fn find_all(&self, conn: &mut DBConn) -> Result<Vec<UserModel>>;
    async fn find_id_by_name(&self, conn: &mut DBConn, name: &str) -> Result<Option<i64>>;
    async fn find_by_name(&self, conn: &mut DBConn, name: &str) -> Result<Option<UserModel>>;
}
