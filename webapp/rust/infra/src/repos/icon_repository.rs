#[cfg(test)]
mod create;
#[cfg(test)]
mod delete_by_user_id;

use crate::qbey_support::{bind_sql_values, bind_qbey_values, SQLValue};
use crate::tables::icon::TABLE_ICONS;
use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::icon::CreateIcon;
use isupipe_core::models::user::UserId;
use isupipe_core::repos::icon_repository::IconRepository;
use qbey::prelude::*;
use qbey_mysql::qbey;

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
        let mut q = qbey(t.table());
        q.and_where(t.user_id().eq(*user_id.inner()));
        q.select(&[t.image()]);
        let (sql, binds) = q.to_sql();
        let image = bind_qbey_values!(sqlx::query_scalar::<_, Vec<u8>>(&sql), binds)
            .fetch_optional(conn)
            .await?;

        Ok(image)
    }

    async fn create(
        &self,
        conn: &mut DBConn,
        icon: &CreateIcon,
    ) -> isupipe_core::repos::Result<i64> {
        let t = &TABLE_ICONS;
        let mut ins = qbey_mysql::qbey_with::<SQLValue>(t.table()).into_insert();
        ins.add_value(&[
            ("user_id", (*icon.user_id.inner()).into()),
            ("image", icon.image.clone().into()),
        ]);
        let (sql, binds) = ins.to_sql();
        let rs = bind_sql_values!(sqlx::query(&sql), binds)
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
        let mut d = qbey(t.table()).into_delete();
        d.and_where(t.user_id().eq(*user_id.inner()));
        let (sql, binds) = d.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(conn)
            .await?;

        Ok(())
    }
}
