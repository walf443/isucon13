use crate::db::DBConn;
use crate::models::user::{CreateUser, User, UserId};
use crate::repos::Result;
use async_trait::async_trait;

#[async_trait]
pub trait UserRepository {
    async fn create(&self, conn: &mut DBConn, user: &CreateUser) -> Result<UserId>;

    fn hash_password(&self, password: &str) -> Result<String> {
        const BCRYPT_DEFAULT_COST: u32 = 4;
        let hashed_password = bcrypt::hash(password, BCRYPT_DEFAULT_COST)?;
        Ok(hashed_password)
    }

    async fn find(&self, conn: &mut DBConn, id: &UserId) -> Result<Option<User>>;
    async fn find_all(&self, conn: &mut DBConn) -> Result<Vec<User>>;
    async fn find_id_by_name(&self, conn: &mut DBConn, name: &str) -> Result<Option<UserId>>;
    async fn find_by_name(&self, conn: &mut DBConn, name: &str) -> Result<Option<User>>;
}

pub trait HaveUserRepository {
    type Repo: UserRepository;

    fn user_repo(&self) -> &Self::Repo;
}
