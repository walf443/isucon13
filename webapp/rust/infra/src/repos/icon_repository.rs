#[cfg(test)]
mod create;
#[cfg(test)]
mod delete_by_user_id;

use crate::tables::icon::IconRow;
use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::icon::{CreateIcon, IconId};
use isupipe_core::models::user::UserId;
use isupipe_core::repos::icon_repository::IconRepository;

#[derive(Clone)]
pub struct IconRepositoryInfra {}

#[async_trait]
impl IconRepository for IconRepositoryInfra {
    async fn find_image_by_user_id<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        user_id: &UserId,
    ) -> isupipe_core::repos::Result<Option<Vec<u8>>> {
        let images = IconRow::filter(IconRow::fields().user_id().eq(user_id))
            .limit(1)
            .select(IconRow::fields().image())
            .exec(conn)
            .await?;

        Ok(images.into_iter().next())
    }

    async fn create<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        icon: &CreateIcon,
    ) -> isupipe_core::repos::Result<IconId> {
        let row = IconRow::create()
            .user_id(&icon.user_id)
            .image(icon.image.clone())
            .exec(conn)
            .await?;

        Ok(row.id)
    }

    async fn delete_by_user_id<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        user_id: &UserId,
    ) -> isupipe_core::repos::Result<()> {
        IconRow::filter(IconRow::fields().user_id().eq(user_id))
            .delete()
            .exec(conn)
            .await?;

        Ok(())
    }
}
