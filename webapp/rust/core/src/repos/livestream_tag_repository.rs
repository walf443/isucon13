use crate::db::DBConn;
use crate::models::livestream::LivestreamId;
use crate::models::livestream_tag::LivestreamTag;
use crate::models::tag::TagId;
use crate::repos::Result;
use async_trait::async_trait;

#[async_trait]
pub trait LivestreamTagRepository {
    async fn insert(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
        tag_id: &TagId,
    ) -> Result<()>;

    async fn find_all_by_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
    ) -> Result<Vec<LivestreamTag>>;

    async fn find_all_by_tag_ids(
        &self,
        conn: &mut DBConn,
        tag_ids: &[TagId],
    ) -> Result<Vec<LivestreamTag>>;
}
