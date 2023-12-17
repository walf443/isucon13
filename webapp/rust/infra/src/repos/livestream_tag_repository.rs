use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::livestream_tag::LivestreamTagModel;
use isupipe_core::repos::livestream_tag_repository::LivestreamTagRepository;

pub struct LivestreamTagRepositoryInfra {}

#[async_trait]
impl LivestreamTagRepository for LivestreamTagRepositoryInfra {
    async fn insert(
        &self,
        conn: &mut DBConn,
        livestream_id: i64,
        tag_id: i64,
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
        livestream_id: i64,
    ) -> isupipe_core::repos::Result<Vec<LivestreamTagModel>> {
        let livestream_tag_models =
            sqlx::query_as("SELECT * FROM livestream_tags WHERE livestream_id = ?")
                .bind(livestream_id)
                .fetch_all(conn)
                .await?;

        Ok(livestream_tag_models)
    }

    async fn find_all_by_tag_ids(
        &self,
        conn: &mut DBConn,
        tag_ids: &Vec<i64>,
    ) -> isupipe_core::repos::Result<Vec<LivestreamTagModel>> {
        let mut query_builder = sqlx::query_builder::QueryBuilder::new(
            "SELECT * FROM livestream_tags WHERE tag_id IN (",
        );
        let mut separated = query_builder.separated(", ");
        for tag_id in tag_ids {
            separated.push_bind(tag_id);
        }
        separated.push_unseparated(") ORDER BY livestream_id DESC");
        let livestreams: Vec<LivestreamTagModel> =
            query_builder.build_query_as().fetch_all(conn).await?;

        Ok(livestreams)
    }
}
