use crate::db::HaveDBPool;
use crate::models::user::{CreateUser, User, UserId};
use crate::repos::theme_repository::{HaveThemeRepository, ThemeRepository};
use crate::repos::user_repository::{HaveUserRepository, UserRepository};
use crate::services::ServiceResult;
use async_trait::async_trait;

#[async_trait]
pub trait UserService {
    async fn create(&self, user: &CreateUser, dark_mode: bool) -> ServiceResult<User>;
    async fn find(&self, id: &UserId) -> ServiceResult<Option<User>>;
    async fn find_by_name(&self, name: &str) -> ServiceResult<Option<User>>;
}

pub trait HaveUserService {
    type Service: UserService;

    fn user_service(&self) -> &Self::Service;
}

pub trait UserServiceImpl: Sync + HaveDBPool + HaveUserRepository + HaveThemeRepository {}

#[async_trait]
impl<T: UserServiceImpl> UserService for T {
    async fn create(&self, user: &CreateUser, dark_mode: bool) -> ServiceResult<User> {
        let mut tx = self.get_db_pool().begin().await?;

        let user_id = self.user_repo().create(&mut tx, user).await?;

        self.theme_repo()
            .create(&mut tx, &user_id, dark_mode)
            .await?;

        tx.commit().await?;

        let hashed_password = self.user_repo().hash_password(&user.password)?;

        Ok(User {
            id: user_id,
            name: user.name.clone(),
            display_name: Some(user.display_name.clone()),
            description: Some(user.description.clone()),
            hashed_password: Some(hashed_password),
        })
    }

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
