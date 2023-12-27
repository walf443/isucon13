use crate::db::DBConn;
use crate::models::icon::CreateIcon;
use crate::models::user::UserId;
use crate::repos::Result;
use async_trait::async_trait;

#[async_trait]
pub trait IconRepository {
    async fn find_image_by_user_id(
        &self,
        conn: &mut DBConn,
        user_id: &UserId,
    ) -> Result<Option<Vec<u8>>>;

    async fn create(&self, conn: &mut DBConn, icon: &CreateIcon) -> Result<i64>;

    async fn delete_by_user_id(&self, conn: &mut DBConn, user_id: &UserId) -> Result<()>;
}

pub trait HaveIconRepository {
    type Repo: IconRepository;

    fn icon_repo(&self) -> &Self::Repo;
}
