use crate::db::HaveDBPool;
use crate::models::theme::Theme;
use crate::models::user::UserId;
use crate::repos::theme_repository::{HaveThemeRepository, ThemeRepository};
use crate::services::ServiceResult;
use async_trait::async_trait;

#[async_trait]
pub trait ThemeService {
    async fn find_by_user_id(&self, user_id: &UserId) -> ServiceResult<Theme>;
}

pub trait HaveThemeService {
    type Service: ThemeService;

    fn theme_service(&self) -> &Self::Service;
}

pub trait ThemeServiceImpl: Sync + HaveDBPool + HaveThemeRepository {}

#[async_trait]
impl<T: ThemeServiceImpl> ThemeService for T {
    async fn find_by_user_id(&self, user_id: &UserId) -> ServiceResult<Theme> {
        let mut conn = self.get_db_pool().acquire().await?;
        let theme = self
            .theme_repo()
            .find_by_user_id(&mut conn, user_id)
            .await?;

        Ok(theme)
    }
}
