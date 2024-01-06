use crate::db::HaveDBPool;
use crate::models::icon::CreateIcon;
use crate::models::user::UserId;
use crate::repos::icon_repository::{HaveIconRepository, IconRepository};
use crate::repos::user_repository::{HaveUserRepository, UserRepository};
use crate::services::ServiceResult;
use async_trait::async_trait;

#[async_trait]
pub trait IconService {
    async fn find_image_by_user_id(&self, user_id: &UserId) -> ServiceResult<Option<Vec<u8>>>;
    async fn find_image_by_user_name(&self, user_name: &str) -> ServiceResult<Option<Vec<u8>>>;
    async fn replace_new_image(&self, user_id: &UserId, image: &[u8]) -> ServiceResult<i64>;
}

pub trait HaveIconService {
    type Service: IconService;

    fn icon_service(&self) -> &Self::Service;
}

pub trait IconServiceImpl: Sync + HaveDBPool + HaveIconRepository + HaveUserRepository {}

#[async_trait]
impl<T: IconServiceImpl> IconService for T {
    async fn find_image_by_user_id(&self, user_id: &UserId) -> ServiceResult<Option<Vec<u8>>> {
        let mut conn = self.get_db_pool().begin().await?;

        let image = self
            .icon_repo()
            .find_image_by_user_id(&mut conn, user_id)
            .await?;

        Ok(image)
    }

    async fn find_image_by_user_name(&self, user_name: &str) -> ServiceResult<Option<Vec<u8>>> {
        let mut conn = self.get_db_pool().begin().await?;

        let user = self
            .user_repo()
            .find_by_name(&mut conn, user_name)
            .await?
            .unwrap();
        let image = self
            .icon_repo()
            .find_image_by_user_id(&mut conn, &user.id)
            .await?;

        Ok(image)
    }

    async fn replace_new_image(&self, user_id: &UserId, image: &[u8]) -> ServiceResult<i64> {
        let mut tx = self.get_db_pool().begin().await?;

        self.icon_repo().delete_by_user_id(&mut tx, user_id).await?;

        let icon_id = self
            .icon_repo()
            .create(
                &mut tx,
                &CreateIcon {
                    user_id: user_id.clone(),
                    image: image.to_vec(),
                },
            )
            .await?;

        tx.commit().await?;

        Ok(icon_id)
    }
}
