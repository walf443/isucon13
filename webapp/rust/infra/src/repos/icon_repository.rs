#[cfg(test)]
mod delete_by_user_id;

use crate::sqipe_support::bind_sqipe_values;
use crate::tables::icon::TABLE_ICONS;
use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::icon::CreateIcon;
use isupipe_core::models::user::UserId;
use isupipe_core::repos::icon_repository::IconRepository;
use sqipe_mysql::sqipe;

#[derive(Clone)]
pub struct IconRepositoryInfra {}

#[async_trait]
impl IconRepository for IconRepositoryInfra {
    async fn find_image_by_user_id(
        &self,
        conn: &mut DBConn,
        user_id: &UserId,
    ) -> isupipe_core::repos::Result<Option<Vec<u8>>> {
        let t = &TABLE_ICONS;
        let mut q = sqipe(t.table_name());
        q.select(&[t.image()]);
        q.and_where(t.user_id().eq(*user_id.inner()));
        let (sql, binds) = q.to_sql();
        let image = bind_sqipe_values!(sqlx::query_scalar::<_, Vec<u8>>(&sql), binds)
            .fetch_optional(conn)
            .await?;

        Ok(image)
    }

    async fn create(
        &self,
        conn: &mut DBConn,
        icon: &CreateIcon,
    ) -> isupipe_core::repos::Result<i64> {
        let rs = sqlx::query("INSERT INTO icons (user_id, image) VALUES (?, ?)")
            .bind(&icon.user_id)
            .bind(&icon.image)
            .execute(conn)
            .await?;
        let icon_id = rs.last_insert_id() as i64;

        Ok(icon_id)
    }

    async fn delete_by_user_id(
        &self,
        conn: &mut DBConn,
        user_id: &UserId,
    ) -> isupipe_core::repos::Result<()> {
        let t = &TABLE_ICONS;
        let mut d = sqipe(t.table_name()).into_delete();
        d.and_where(t.user_id().eq(*user_id.inner()));
        let (sql, binds) = d.to_sql();
        bind_sqipe_values!(sqlx::query(&sql), binds)
            .execute(conn)
            .await?;

        Ok(())
    }
}
