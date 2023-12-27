use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::livestream_tag::LivestreamTag;
use isupipe_core::models::tag::TagId;
use isupipe_core::repos::livestream_tag_repository::LivestreamTagRepository;

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
        tag_ids: &[TagId],
    ) -> isupipe_core::repos::Result<Vec<LivestreamTag>> {
        let mut query_builder = sqlx::query_builder::QueryBuilder::new(
            "SELECT * FROM livestream_tags WHERE tag_id IN (",
        );
        let mut separated = query_builder.separated(", ");
        for tag_id in tag_ids {
            separated.push_bind(tag_id);
        }
        separated.push_unseparated(") ORDER BY livestream_id DESC");
        let livestreams: Vec<LivestreamTag> =
            query_builder.build_query_as().fetch_all(conn).await?;

        Ok(livestreams)
    }
}
