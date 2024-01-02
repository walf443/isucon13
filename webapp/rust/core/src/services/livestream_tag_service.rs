use crate::db::HaveDBPool;
use crate::models::livestream::LivestreamId;
use crate::models::livestream_tag::LivestreamTag;
use crate::repos::livestream_tag_repository::{
    HaveLivestreamTagRepository, LivestreamTagRepository,
};
use crate::services::ServiceResult;
use async_trait::async_trait;

#[async_trait]
pub trait LivestreamTagService {
    async fn find_all_by_livestream_id(
        &self,
        livestream_id: &LivestreamId,
    ) -> ServiceResult<Vec<LivestreamTag>>;
}

pub trait HaveLivestreamTagService {
    type Service: LivestreamTagService;

    fn livestream_tag_service(&self) -> &Self::Service;
}

pub trait LivestreamTagServiceImpl: Sync + HaveDBPool + HaveLivestreamTagRepository {}

#[async_trait]
impl<T: LivestreamTagServiceImpl> LivestreamTagService for T {
    async fn find_all_by_livestream_id(
        &self,
        livestream_id: &LivestreamId,
    ) -> ServiceResult<Vec<LivestreamTag>> {
        let mut conn = self.get_db_pool().acquire().await?;
        let livestream_tags = self
            .livestream_tag_repo()
            .find_all_by_livestream_id(&mut conn, livestream_id)
            .await?;
        Ok(livestream_tags)
    }
}
