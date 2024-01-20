#[cfg(test)]
mod find;
#[cfg(test)]
mod find_all;
#[cfg(test)]
mod find_ids_by_name;

use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::tag::{Tag, TagId, TagName};
use isupipe_core::repos::tag_repository::TagRepository;

#[derive(Clone)]
pub struct TagRepositoryInfra {}

#[async_trait]
impl TagRepository for TagRepositoryInfra {
    async fn find(&self, conn: &mut DBConn, id: &TagId) -> isupipe_core::repos::Result<Tag> {
        let tag_model = sqlx::query_as!(Tag, "SELECT * FROM tags WHERE id = ?", id)
            .fetch_one(conn)
            .await?;

        Ok(tag_model)
    }

    async fn find_all(&self, conn: &mut DBConn) -> isupipe_core::repos::Result<Vec<Tag>> {
        let tag_models = sqlx::query_as!(Tag, "SELECT * FROM tags")
            .fetch_all(conn)
            .await?;

        Ok(tag_models)
    }

    async fn find_ids_by_name(
        &self,
        conn: &mut DBConn,
        name: &TagName,
    ) -> isupipe_core::repos::Result<Vec<TagId>> {
        let tag_id_list =
            sqlx::query_scalar!("SELECT id as `id:TagId` FROM tags WHERE name = ?", name)
                .fetch_all(conn)
                .await?;

        Ok(tag_id_list)
    }
}
