use crate::db::HaveDBPool;
use crate::models::user::{User, UserId};
use crate::repos::user_repository::{HaveUserRepository, UserRepository};
use crate::services::ServiceResult;
use async_trait::async_trait;

#[async_trait]
pub trait UserService {
    async fn find(&self, id: &UserId) -> ServiceResult<Option<User>>;
    async fn find_by_name(&self, name: &str) -> ServiceResult<Option<User>>;
}

pub trait HaveUserService {
    type Service: UserService;

    fn user_service(&self) -> &Self::Service;
}

pub trait UserServiceImpl: Sync + HaveDBPool + HaveUserRepository {}

#[async_trait]
impl<T: UserServiceImpl> UserService for T {
    async fn find(&self, id: &UserId) -> ServiceResult<Option<User>> {
        let mut conn = self.get_db_pool().acquire().await?;
        let user = self.user_repo().find(&mut conn, id).await?;
        Ok(user)
    }

    async fn find_by_name(&self, name: &str) -> ServiceResult<Option<User>> {
        let mut conn = self.get_db_pool().acquire().await?;
        let user = self.user_repo().find_by_name(&mut conn, name).await?;
        Ok(user)
    }
}
