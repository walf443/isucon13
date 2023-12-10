use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::livestream::LivestreamModel;
use isupipe_core::repos::livestream_repository::LivestreamRepository;

pub struct LivestreamRepositoryInfra {}

#[async_trait]
impl LivestreamRepository for LivestreamRepositoryInfra {
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
}
