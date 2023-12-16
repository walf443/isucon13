use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::livestream::LivestreamModel;
use isupipe_core::repos::livestream_repository::LivestreamRepository;

pub struct LivestreamRepositoryInfra {}

#[async_trait]
impl LivestreamRepository for LivestreamRepositoryInfra {
    async fn find_all(
        &self,
        conn: &mut DBConn,
    ) -> isupipe_core::repos::Result<Vec<LivestreamModel>> {
        let livestreams: Vec<LivestreamModel> = sqlx::query_as("SELECT * FROM livestreams")
            .fetch_all(conn)
            .await?;

        Ok(livestreams)
    }

    async fn find_all_by_user_id(
        &self,
        conn: &mut DBConn,
        user_id: i64,
    ) -> isupipe_core::repos::Result<Vec<LivestreamModel>> {
        let livestream_models: Vec<LivestreamModel> =
            sqlx::query_as("SELECT * FROM livestreams WHERE user_id = ?")
                .bind(user_id)
                .fetch_all(conn)
                .await?;

        Ok(livestream_models)
    }

    async fn find(
        &self,
        conn: &mut DBConn,
        id: i64,
    ) -> isupipe_core::repos::Result<Option<LivestreamModel>> {
        let livestream = sqlx::query_as("SELECT * FROM livestreams WHERE id = ?")
            .bind(id)
            .fetch_optional(conn)
            .await?;

        Ok(livestream)
    }

    async fn exist_by_id_and_user_id(
        &self,
        conn: &mut DBConn,
        id: i64,
        user_id: i64,
    ) -> isupipe_core::repos::Result<bool> {
        let livestreams: Vec<LivestreamModel> =
            sqlx::query_as("SELECT * FROM livestreams WHERE id = ? AND user_id = ?")
                .bind(id)
                .bind(user_id)
                .fetch_all(conn)
                .await?;

        Ok(!livestreams.is_empty())
    }
}
