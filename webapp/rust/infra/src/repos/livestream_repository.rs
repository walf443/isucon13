use crate::sqipe_support::bind_sqipe_values;
use crate::tables::livestream::TABLE_LIVESTREAMS;
use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::livestream::{CreateLivestream, Livestream, LivestreamId};
use isupipe_core::models::user::UserId;
use isupipe_core::repos::livestream_repository::LivestreamRepository;
use sqipe_mysql::sqipe;

#[derive(Clone)]
pub struct LivestreamRepositoryInfra {}

#[async_trait]
impl LivestreamRepository for LivestreamRepositoryInfra {
    async fn create(
        &self,
        conn: &mut DBConn,
        stream: &CreateLivestream,
    ) -> isupipe_core::repos::Result<LivestreamId> {
        let rs = sqlx::query("INSERT INTO livestreams (user_id, title, description, playlist_url, thumbnail_url, start_at, end_at) VALUES(?, ?, ?, ?, ?, ?, ?)")
            .bind(&stream.user_id)
            .bind(&stream.title)
            .bind(&stream.description)
            .bind(&stream.playlist_url)
            .bind(&stream.thumbnail_url)
            .bind(stream.start_at)
            .bind(stream.end_at)
            .execute(conn)
            .await?;

        let livestream_id = rs.last_insert_id() as i64;
        Ok(LivestreamId::new(livestream_id))
    }

    async fn find_all(&self, conn: &mut DBConn) -> isupipe_core::repos::Result<Vec<Livestream>> {
        let t = &TABLE_LIVESTREAMS;
        let q = sqipe(t.table_name());
        let (sql, binds) = q.to_sql();
        let livestreams = bind_sqipe_values!(sqlx::query_as::<_, Livestream>(&sql), binds)
            .fetch_all(conn)
            .await?;

        Ok(livestreams)
    }

    async fn find_all_order_by_id_desc(
        &self,
        conn: &mut DBConn,
    ) -> isupipe_core::repos::Result<Vec<Livestream>> {
        let t = &TABLE_LIVESTREAMS;
        let mut q = sqipe(t.table_name());
        q.order_by(t.id().desc());
        let (sql, binds) = q.to_sql();
        let livestreams = bind_sqipe_values!(sqlx::query_as::<_, Livestream>(&sql), binds)
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
        let mut q = sqipe(t.table_name());
        q.order_by(t.id().desc());
        q.limit(limit as u64);
        let (sql, binds) = q.to_sql();
        let livestreams = bind_sqipe_values!(sqlx::query_as::<_, Livestream>(&sql), binds)
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
        let mut q = sqipe(t.table_name());
        q.and_where(("user_id", *user_id.inner()));
        let (sql, binds) = q.to_sql();
        let livestream_models = bind_sqipe_values!(sqlx::query_as::<_, Livestream>(&sql), binds)
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
        let mut q = sqipe(t.table_name());
        q.and_where(("id", *id.inner()));
        let (sql, binds) = q.to_sql();
        let livestream = bind_sqipe_values!(sqlx::query_as::<_, Livestream>(&sql), binds)
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
        let mut q = sqipe(t.table_name());
        q.and_where(("id", *id.inner()));
        q.and_where(("user_id", *user_id.inner()));
        let (sql, binds) = q.to_sql();
        let livestreams: Vec<Livestream> =
            bind_sqipe_values!(sqlx::query_as::<_, Livestream>(&sql), binds)
                .fetch_all(conn)
                .await?;

        Ok(!livestreams.is_empty())
    }
}

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
#[cfg(test)]
mod exist_by_id_and_user_id;
