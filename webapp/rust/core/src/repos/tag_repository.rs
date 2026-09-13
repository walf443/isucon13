use crate::db::DBConn;
use crate::models::tag::{Tag, TagId, TagName};
use crate::repos::Result;
use async_trait::async_trait;

#[cfg_attr(any(feature = "test", test), mockall::automock)]
#[async_trait]
pub trait TagRepository {
    async fn find<'c>(&self, conn: &'c mut DBConn<'c>, id: &TagId) -> Result<Tag>;
    async fn find_all<'c>(&self, conn: &'c mut DBConn<'c>) -> Result<Vec<Tag>>;

    async fn find_ids_by_name<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        name: &TagName,
    ) -> Result<Vec<TagId>>;
}

pub trait HaveTagRepository {
    type Repo: Sync + TagRepository;

    fn tag_repo(&self) -> &Self::Repo;
}
