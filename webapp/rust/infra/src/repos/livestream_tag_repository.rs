use crate::sqipe_support::bind_sqipe_values;
use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::livestream_tag::LivestreamTag;
use isupipe_core::models::tag::TagId;
use isupipe_core::repos::livestream_tag_repository::LivestreamTagRepository;
use sqipe::col;
use sqipe_mysql::sqipe;

#[derive(Clone)]
pub struct LivestreamTagRepositoryInfra {}

#[async_trait]
impl LivestreamTagRepository for LivestreamTagRepositoryInfra {
    async fn insert(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
        tag_id: &TagId,
    ) -> isupipe_core::repos::Result<()> {
        sqlx::query("INSERT INTO livestream_tags (livestream_id, tag_id) VALUES (?, ?)")
            .bind(livestream_id)
            .bind(tag_id)
            .execute(conn)
            .await?;

        Ok(())
    }

    async fn find_all_by_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
    ) -> isupipe_core::repos::Result<Vec<LivestreamTag>> {
        let mut q = sqipe("livestream_tags");
        q.and_where(("livestream_id", *livestream_id.inner()));
        let (sql, binds) = q.to_sql();
        let livestream_tag_models =
            bind_sqipe_values!(sqlx::query_as::<_, LivestreamTag>(&sql), binds)
                .fetch_all(conn)
                .await?;

        Ok(livestream_tag_models)
    }

    async fn find_all_by_tag_ids(
        &self,
        conn: &mut DBConn,
        tag_ids: &[TagId],
    ) -> isupipe_core::repos::Result<Vec<LivestreamTag>> {
        let mut q = sqipe("livestream_tags");
        let id_values: Vec<sqipe::Value> =
            tag_ids.iter().map(|id| sqipe::Value::Int(*id.inner())).collect();
        q.and_where(col("tag_id").included(id_values.as_slice()));
        q.order_by(col("livestream_id").desc());
        let (sql, binds) = q.to_sql();
        let livestreams =
            bind_sqipe_values!(sqlx::query_as::<_, LivestreamTag>(&sql), binds)
                .fetch_all(conn)
                .await?;

        Ok(livestreams)
    }
}
