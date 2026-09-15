use crate::qbey_support::bind_qbey_values;
use crate::tables::livestream::TABLE_LIVESTREAMS;
use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::livestream::{CreateLivestream, Livestream, LivestreamId};
use isupipe_core::models::user::UserId;
use isupipe_core::repos::livestream_repository::LivestreamRepository;
use qbey::prelude::*;
use qbey_mysql::qbey;

struct InsertLivestream<'a>(&'a CreateLivestream);

impl qbey::ToInsertRow<qbey::Value, String> for InsertLivestream<'_> {
    fn to_insert_row(&self) -> Vec<(String, qbey::Value)> {
        let mut row = TABLE_LIVESTREAMS.row();
        row.user_id(&self.0.user_id)
            .title(&self.0.title)
            .description(&self.0.description)
            .playlist_url(&self.0.playlist_url)
            .thumbnail_url(&self.0.thumbnail_url)
            .start_at(self.0.start_at)
            .end_at(self.0.end_at);
        row.to_insert_row()
    }
}

#[derive(Clone)]
pub struct LivestreamRepositoryInfra {}

#[async_trait]
impl LivestreamRepository for LivestreamRepositoryInfra {
    async fn create(
        &self,
        conn: &mut DBConn,
        stream: &CreateLivestream,
    ) -> isupipe_core::repos::Result<LivestreamId> {
        let t = &TABLE_LIVESTREAMS;
        let mut ins = qbey(t.table()).into_insert();
        ins.add_value(&InsertLivestream(stream));
        let (sql, binds) = ins.into_sql();
        let rs = bind_qbey_values!(sqlx::query(sqlx::AssertSqlSafe(sql)), binds)
            .execute(conn)
            .await?;

        let livestream_id = rs.last_insert_id() as i64;
        Ok(LivestreamId::new(livestream_id))
    }

    async fn find_all(&self, conn: &mut DBConn) -> isupipe_core::repos::Result<Vec<Livestream>> {
        let t = &TABLE_LIVESTREAMS;
        let q = qbey(t.table());
        let (sql, binds) = q.into_sql();
        let livestreams = bind_qbey_values!(
            sqlx::query_as::<_, Livestream>(sqlx::AssertSqlSafe(sql)),
            binds
        )
        .fetch_all(conn)
        .await?;

        Ok(livestreams)
    }

    async fn find_all_order_by_id_desc(
        &self,
        conn: &mut DBConn,
    ) -> isupipe_core::repos::Result<Vec<Livestream>> {
        let t = &TABLE_LIVESTREAMS;
        let mut q = qbey(t.table());
        q.order_by(t.id().desc());
        let (sql, binds) = q.into_sql();
        let livestreams = bind_qbey_values!(
            sqlx::query_as::<_, Livestream>(sqlx::AssertSqlSafe(sql)),
            binds
        )
        .fetch_all(conn)
        .await?;

        Ok(livestreams)
    }

    async fn find_all_order_by_id_desc_limit(
        &self,
        conn: &mut DBConn,
        limit: i64,
    ) -> isupipe_core::repos::Result<Vec<Livestream>> {
        let t = &TABLE_LIVESTREAMS;
        let mut q = qbey(t.table());
        q.order_by(t.id().desc());
        q.limit(limit as u64);
        let (sql, binds) = q.into_sql();
        let livestreams = bind_qbey_values!(
            sqlx::query_as::<_, Livestream>(sqlx::AssertSqlSafe(sql)),
            binds
        )
        .fetch_all(conn)
        .await?;

        Ok(livestreams)
    }

    async fn find_all_by_user_id(
        &self,
        conn: &mut DBConn,
        user_id: &UserId,
    ) -> isupipe_core::repos::Result<Vec<Livestream>> {
        let t = &TABLE_LIVESTREAMS;
        let mut q = qbey(t.table());
        q.and_where(t.user_id().eq(user_id));
        let (sql, binds) = q.into_sql();
        let livestream_models = bind_qbey_values!(
            sqlx::query_as::<_, Livestream>(sqlx::AssertSqlSafe(sql)),
            binds
        )
        .fetch_all(conn)
        .await?;

        Ok(livestream_models)
    }

    async fn find(
        &self,
        conn: &mut DBConn,
        id: &LivestreamId,
    ) -> isupipe_core::repos::Result<Option<Livestream>> {
        let t = &TABLE_LIVESTREAMS;
        let mut q = qbey(t.table());
        q.and_where(t.id().eq(id));
        let (sql, binds) = q.into_sql();
        let livestream = bind_qbey_values!(
            sqlx::query_as::<_, Livestream>(sqlx::AssertSqlSafe(sql)),
            binds
        )
        .fetch_optional(conn)
        .await?;

        Ok(livestream)
    }

    async fn exist_by_id_and_user_id(
        &self,
        conn: &mut DBConn,
        id: &LivestreamId,
        user_id: &UserId,
    ) -> isupipe_core::repos::Result<bool> {
        let t = &TABLE_LIVESTREAMS;
        let mut q = qbey(t.table());
        q.and_where(t.id().eq(id));
        q.and_where(t.user_id().eq(user_id));
        let (sql, binds) = q.into_sql();
        let livestreams: Vec<Livestream> = bind_qbey_values!(
            sqlx::query_as::<_, Livestream>(sqlx::AssertSqlSafe(sql)),
            binds
        )
        .fetch_all(conn)
        .await?;

        Ok(!livestreams.is_empty())
    }
}

#[cfg(test)]
mod create;
#[cfg(test)]
mod exist_by_id_and_user_id;
#[cfg(test)]
mod find;
#[cfg(test)]
mod find_all;
#[cfg(test)]
mod find_all_by_user_id;
#[cfg(test)]
mod find_all_order_by_id_desc;
#[cfg(test)]
mod find_all_order_by_id_desc_limit;
