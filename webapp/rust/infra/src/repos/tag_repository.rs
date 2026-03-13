#[cfg(test)]
mod find;
#[cfg(test)]
mod find_all;
#[cfg(test)]
mod find_ids_by_name;

use crate::sqipe_support::bind_sqipe_values;
use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::tag::{Tag, TagId, TagName};
use isupipe_core::repos::tag_repository::TagRepository;
use sqipe_mysql::sqipe;

#[derive(Clone)]
pub struct TagRepositoryInfra {}

#[async_trait]
impl TagRepository for TagRepositoryInfra {
    async fn find(&self, conn: &mut DBConn, id: &TagId) -> isupipe_core::repos::Result<Tag> {
        let mut q = sqipe("tags");
        q.and_where(("id", *id.inner()));

        let (sql, binds) = q.to_sql();
        let tag_model = bind_sqipe_values!(sqlx::query_as::<_, Tag>(&sql), binds)
            .fetch_one(conn)
            .await?;

        Ok(tag_model)
    }

    async fn find_all(&self, conn: &mut DBConn) -> isupipe_core::repos::Result<Vec<Tag>> {
        let q = sqipe("tags");

        let (sql, binds) = q.to_sql();
        let tag_models = bind_sqipe_values!(sqlx::query_as::<_, Tag>(&sql), binds)
            .fetch_all(conn)
            .await?;

        Ok(tag_models)
    }

    async fn find_ids_by_name(
        &self,
        conn: &mut DBConn,
        name: &TagName,
    ) -> isupipe_core::repos::Result<Vec<TagId>> {
        let mut q = sqipe("tags");
        q.select(&["id"]);
        q.and_where(("name", name.inner().clone()));

        let (sql, binds) = q.to_sql();
        let tag_id_list = bind_sqipe_values!(sqlx::query_scalar::<_, TagId>(&sql), binds)
            .fetch_all(conn)
            .await?;

        Ok(tag_id_list)
    }
}
