use async_session::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::tag::TagModel;
use isupipe_core::repos::tag_repository::TagRepository;

pub struct TagRepositoryInfra {}

#[async_trait]
impl TagRepository for TagRepositoryInfra {
    async fn find_all(&self, conn: &mut DBConn) -> isupipe_core::repos::Result<Vec<TagModel>> {
        let tag_models: Vec<TagModel> =
            sqlx::query_as("SELECT * FROM tags").fetch_all(conn).await?;

        Ok(tag_models)
    }
}
