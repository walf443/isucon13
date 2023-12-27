use crate::db::HaveDBPool;
use crate::repos::icon_repository::{HaveIconRepository, IconRepository};
use crate::repos::user_repository::{HaveUserRepository, UserRepository};
use crate::services::ServiceResult;
use async_trait::async_trait;

#[async_trait]
pub trait UserIconService {
    async fn find_image_by_user_name(&self, user_name: &str) -> ServiceResult<Option<Vec<u8>>>;
}

pub trait HaveUserIconService {
    type Service: UserIconService;

    fn user_icon_service(&self) -> &Self::Service;
}

pub trait UserIconServiceImpl: Sync + HaveDBPool + HaveIconRepository + HaveUserRepository {}

#[async_trait]
impl<T: UserIconServiceImpl> UserIconService for T {
    async fn find_image_by_user_name(&self, user_name: &str) -> ServiceResult<Option<Vec<u8>>> {
        let mut conn = self.get_db_pool().begin().await?;

        let user = self
            .user_repo()
            .find_by_name(&mut *conn, user_name)
            .await?
            .unwrap();
        let image = self
            .icon_repo()
            .find_image_by_user_id(&mut *conn, &user.id)
            .await?;

        Ok(image)
    }
}
