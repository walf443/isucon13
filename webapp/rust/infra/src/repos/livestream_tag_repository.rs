#[cfg(test)]
mod find_all_by_livestream_id;
#[cfg(test)]
mod find_all_by_tag_ids;
#[cfg(test)]
mod insert;

use crate::qbey_support::bind_qbey_values;
use crate::tables::livestream_tag::TABLE_LIVESTREAM_TAGS;
use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::livestream_tag::LivestreamTag;
use isupipe_core::models::tag::TagId;
use isupipe_core::repos::livestream_tag_repository::LivestreamTagRepository;
use qbey::prelude::*;
use qbey_mysql::qbey;

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
        let t = &TABLE_LIVESTREAM_TAGS;
        let mut ins = qbey(t.table()).into_insert();
        ins.add_value(&[
            ("livestream_id", (*livestream_id.inner()).into()),
            ("tag_id", (*tag_id.inner()).into()),
        ]);
        let (sql, binds) = ins.into_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(conn)
            .await?;

        Ok(())
    }

    async fn find_all_by_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
    ) -> isupipe_core::repos::Result<Vec<LivestreamTag>> {
        let t = &TABLE_LIVESTREAM_TAGS;
        let mut q = qbey(t.table());
        q.and_where(t.livestream_id().eq(*livestream_id.inner()));
        let (sql, binds) = q.into_sql();
        let livestream_tag_models =
            bind_qbey_values!(sqlx::query_as::<_, LivestreamTag>(&sql), binds)
                .fetch_all(conn)
                .await?;

        Ok(livestream_tag_models)
    }

    async fn find_all_by_tag_ids(
        &self,
        conn: &mut DBConn,
        tag_ids: &[TagId],
    ) -> isupipe_core::repos::Result<Vec<LivestreamTag>> {
        let t = &TABLE_LIVESTREAM_TAGS;
        let mut q = qbey(t.table());
        let id_values: Vec<qbey::Value> = tag_ids
            .iter()
            .map(|id| qbey::Value::Int(*id.inner()))
            .collect();
        q.and_where(t.tag_id().included(id_values.as_slice()));
        q.order_by(t.livestream_id().desc());
        let (sql, binds) = q.into_sql();
        let livestreams = bind_qbey_values!(sqlx::query_as::<_, LivestreamTag>(&sql), binds)
            .fetch_all(conn)
            .await?;

        Ok(livestreams)
    }
}
