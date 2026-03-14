#[cfg(test)]
mod create;
#[cfg(test)]
mod find_by_user_id;

use crate::sqipe_support::bind_sqipe_values;
use crate::tables::theme::TABLE_THEMES;
use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::theme::Theme;
use isupipe_core::models::user::UserId;
use isupipe_core::repos::theme_repository::ThemeRepository;
use sqipe::IntoColRef;
use sqipe_mysql::sqipe;

#[derive(Clone)]
pub struct ThemeRepositoryInfra {}

#[async_trait]
impl ThemeRepository for ThemeRepositoryInfra {
    async fn create(
        &self,
        conn: &mut DBConn,
        user_id: &UserId,
        dark_mode: bool,
    ) -> isupipe_core::repos::Result<()> {
        sqlx::query("INSERT INTO themes (user_id, dark_mode) VALUES(?, ?)")
            .bind(user_id)
            .bind(dark_mode)
            .execute(conn)
            .await?;

        Ok(())
    }

    async fn find_by_user_id(
        &self,
        conn: &mut DBConn,
        user_id: &UserId,
    ) -> isupipe_core::repos::Result<Theme> {
        let t = &TABLE_THEMES;
        let mut q = sqipe(t.table_name());
        q.select_cols(&[
            t.id().into_col_ref(),
            t.user_id().into_col_ref(),
            t.dark_mode().into_col_ref(),
        ]);
        q.and_where(("user_id", *user_id.inner()));
        let (sql, binds) = q.to_sql();
        let theme_model = bind_sqipe_values!(sqlx::query_as::<_, Theme>(&sql), binds)
            .fetch_one(conn)
            .await?;

        Ok(theme_model)
    }
}
