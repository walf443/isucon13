use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::livestream::{CreateLivestream, Livestream, LivestreamId};
use isupipe_core::models::user::UserId;
use isupipe_core::repos::livestream_repository::LivestreamRepository;

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
        let livestreams: Vec<Livestream> = sqlx::query_as("SELECT * FROM livestreams")
            .fetch_all(conn)
            .await?;

        Ok(livestreams)
    }

    async fn find_all_order_by_id_desc(
        &self,
        conn: &mut DBConn,
    ) -> isupipe_core::repos::Result<Vec<Livestream>> {
        let livestreams: Vec<Livestream> =
            sqlx::query_as("SELECT * FROM livestreams ORDER BY id DESC")
                .fetch_all(conn)
                .await?;

        Ok(livestreams)
    }

    async fn find_all_order_by_id_desc_limit(
        &self,
        conn: &mut DBConn,
        limit: i64,
    ) -> isupipe_core::repos::Result<Vec<Livestream>> {
        let livestreams: Vec<Livestream> =
            sqlx::query_as("SELECT * FROM livestreams ORDER BY id DESC LIMIT ?")
                .bind(limit)
                .fetch_all(conn)
                .await?;

        Ok(livestreams)
    }

    async fn find_all_by_user_id(
        &self,
        conn: &mut DBConn,
        user_id: &UserId,
    ) -> isupipe_core::repos::Result<Vec<Livestream>> {
        let livestream_models: Vec<Livestream> =
            sqlx::query_as("SELECT * FROM livestreams WHERE user_id = ?")
                .bind(user_id)
                .fetch_all(conn)
                .await?;

        Ok(livestream_models)
    }

    async fn find(
        &self,
        conn: &mut DBConn,
        id: &LivestreamId,
    ) -> isupipe_core::repos::Result<Option<Livestream>> {
        let livestream = sqlx::query_as("SELECT * FROM livestreams WHERE id = ?")
            .bind(id)
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
        let livestreams: Vec<Livestream> =
            sqlx::query_as("SELECT * FROM livestreams WHERE id = ? AND user_id = ?")
                .bind(id)
                .bind(user_id)
                .fetch_all(conn)
                .await?;

        Ok(!livestreams.is_empty())
    }
}
