use crate::db::DBConn;
use crate::models::livestream_tag::LivestreamTagModel;
use crate::repos::Result;
use async_trait::async_trait;

#[async_trait]
pub trait LivestreamTagRepository {
    async fn find_all_by_tag_ids(
        &self,
        conn: &mut DBConn,
        tag_ids: &Vec<i64>,
    ) -> Result<Vec<LivestreamTagModel>>;
}
