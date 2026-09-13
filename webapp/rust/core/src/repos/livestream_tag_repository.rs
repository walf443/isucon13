use crate::db::DBConn;
use crate::models::livestream::LivestreamId;
use crate::models::livestream_tag::LivestreamTag;
use crate::models::tag::TagId;
use crate::repos::Result;
use async_trait::async_trait;

#[cfg_attr(any(feature = "test", test), mockall::automock)]
#[async_trait]
pub trait LivestreamTagRepository {
    async fn insert<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        livestream_id: &LivestreamId,
        tag_id: &TagId,
    ) -> Result<()>;

    async fn find_all_by_livestream_id<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        livestream_id: &LivestreamId,
    ) -> Result<Vec<LivestreamTag>>;

    async fn find_all_by_tag_ids<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        tag_ids: &[TagId],
    ) -> Result<Vec<LivestreamTag>>;
}

pub trait HaveLivestreamTagRepository {
    type Repo: Sync + LivestreamTagRepository;

    fn livestream_tag_repo(&self) -> &Self::Repo;
}
