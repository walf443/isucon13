#[cfg(test)]
mod find;
#[cfg(test)]
mod find_all;
#[cfg(test)]
mod find_ids_by_name;

use crate::qbey_support::bind_qbey_values;
use crate::tables::tag::TABLE_TAGS;
use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::tag::{Tag, TagId, TagName};
use isupipe_core::repos::tag_repository::TagRepository;
use qbey_mysql::qbey;

#[derive(Clone)]
pub struct TagRepositoryInfra {}

#[async_trait]
impl TagRepository for TagRepositoryInfra {
    async fn find(&self, conn: &mut DBConn, id: &TagId) -> isupipe_core::repos::Result<Tag> {
        let t = &TABLE_TAGS;
        let mut q = qbey(t.table_name());
        q.and_where(t.id().eq(*id.inner()));

        let (sql, binds) = q.to_sql();
        let tag_model = bind_qbey_values!(sqlx::query_as::<_, Tag>(&sql), binds)
            .fetch_one(conn)
            .await?;

        Ok(tag_model)
    }

    async fn find_all(&self, conn: &mut DBConn) -> isupipe_core::repos::Result<Vec<Tag>> {
        let t = &TABLE_TAGS;
        let q = qbey(t.table_name());

        let (sql, binds) = q.to_sql();
        let tag_models = bind_qbey_values!(sqlx::query_as::<_, Tag>(&sql), binds)
            .fetch_all(conn)
            .await?;

        Ok(tag_models)
    }

    async fn find_ids_by_name(
        &self,
        conn: &mut DBConn,
        name: &TagName,
    ) -> isupipe_core::repos::Result<Vec<TagId>> {
        let t = &TABLE_TAGS;
        let mut q = qbey(t.table_name());
        q.select(&[t.id()]);
        q.and_where(t.name().eq(name.inner().clone()));

        let (sql, binds) = q.to_sql();
        let tag_id_list = bind_qbey_values!(sqlx::query_scalar::<_, TagId>(&sql), binds)
            .fetch_all(conn)
            .await?;

        Ok(tag_id_list)
    }
}
