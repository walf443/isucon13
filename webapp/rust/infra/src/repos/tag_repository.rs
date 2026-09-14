#[cfg(test)]
mod find;
#[cfg(test)]
mod find_all;
#[cfg(test)]
mod find_ids_by_name;

use crate::tables::tag::TagRow;
use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::tag::{Tag, TagId, TagName};
use isupipe_core::repos::tag_repository::TagRepository;

#[derive(Clone)]
pub struct TagRepositoryInfra {}

#[async_trait]
impl TagRepository for TagRepositoryInfra {
    async fn find<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        id: &TagId,
    ) -> isupipe_core::repos::Result<Tag> {
        let row = TagRow::filter(TagRow::fields().id().eq(id))
            .one()
            .exec(conn)
            .await?;

        Ok(row.into())
    }

    async fn find_all<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
    ) -> isupipe_core::repos::Result<Vec<Tag>> {
        let rows = TagRow::all().exec(conn).await?;

        Ok(rows.into_iter().map(Into::into).collect())
    }

    async fn find_ids_by_name<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        name: &TagName,
    ) -> isupipe_core::repos::Result<Vec<TagId>> {
        let ids = TagRow::filter(TagRow::fields().name().eq(name))
            .select(TagRow::fields().id())
            .exec(conn)
            .await?;

        Ok(ids)
    }
}
