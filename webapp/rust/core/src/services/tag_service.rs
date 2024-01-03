use crate::db::HaveDBPool;
use crate::models::tag::{Tag, TagId};
use crate::repos::tag_repository::{HaveTagRepository, TagRepository};
use crate::services::ServiceResult;
use async_trait::async_trait;

#[async_trait]
pub trait TagService {
    async fn find(&self, tag_id: &TagId) -> ServiceResult<Tag>;
    async fn find_all(&self) -> ServiceResult<Vec<Tag>>;
}

pub trait HaveTagService {
    type Service: Sync + TagService;

    fn tag_service(&self) -> &Self::Service;
}

pub trait TagServiceImpl: Sync + HaveDBPool + HaveTagRepository {}

#[async_trait]
impl<T: TagServiceImpl> TagService for T {
    async fn find(&self, tag_id: &TagId) -> ServiceResult<Tag> {
        let mut conn = self.get_db_pool().acquire().await?;
        let tag = self.tag_repo().find(&mut conn, tag_id).await?;
        Ok(tag)
    }

    async fn find_all(&self) -> ServiceResult<Vec<Tag>> {
        let mut conn = self.get_db_pool().acquire().await?;
        let tags = self.tag_repo().find_all(&mut conn).await?;

        Ok(tags)
    }
}
