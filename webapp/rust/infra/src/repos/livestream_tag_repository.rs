#[cfg(test)]
mod find_all_by_livestream_id;
#[cfg(test)]
mod find_all_by_tag_ids;
#[cfg(test)]
mod insert;

use crate::tables::livestream_tag::LivestreamTagRow;
use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::livestream_tag::LivestreamTag;
use isupipe_core::models::tag::TagId;
use isupipe_core::repos::livestream_tag_repository::LivestreamTagRepository;

#[derive(Clone)]
pub struct LivestreamTagRepositoryInfra {}

#[async_trait]
impl LivestreamTagRepository for LivestreamTagRepositoryInfra {
    async fn insert<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        livestream_id: &LivestreamId,
        tag_id: &TagId,
    ) -> isupipe_core::repos::Result<()> {
        LivestreamTagRow::create()
            .livestream_id(livestream_id.inner())
            .tag_id(tag_id.inner())
            .exec(conn)
            .await?;

        Ok(())
    }

    async fn find_all_by_livestream_id<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        livestream_id: &LivestreamId,
    ) -> isupipe_core::repos::Result<Vec<LivestreamTag>> {
        let rows = LivestreamTagRow::filter(
            LivestreamTagRow::fields()
                .livestream_id()
                .eq(livestream_id.inner()),
        )
        .exec(conn)
        .await?;

        Ok(rows.into_iter().map(Into::into).collect())
    }

    async fn find_all_by_tag_ids<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        tag_ids: &[TagId],
    ) -> isupipe_core::repos::Result<Vec<LivestreamTag>> {
        let ids: Vec<i64> = tag_ids.iter().map(|id| *id.inner()).collect();
        let rows = LivestreamTagRow::filter(LivestreamTagRow::fields().tag_id().in_list(ids))
            .order_by(LivestreamTagRow::fields().livestream_id().desc())
            .exec(conn)
            .await?;

        Ok(rows.into_iter().map(Into::into).collect())
    }
}
