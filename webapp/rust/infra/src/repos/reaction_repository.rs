use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::reaction::ReactionModel;
use isupipe_core::repos::reaction_repository::ReactionRepository;

pub struct ReactionRepositoryInfra {}

#[async_trait]
impl ReactionRepository for ReactionRepositoryInfra {
    async fn find_all_by_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: i64,
    ) -> isupipe_core::repos::Result<Vec<ReactionModel>> {
        let reaction_models: Vec<ReactionModel> = sqlx::query_as(
            "SELECT * FROM reactions WHERE livestream_id = ? ORDER BY created_at DESC",
        )
        .bind(livestream_id)
        .fetch_all(conn)
        .await?;

        Ok(reaction_models)
    }
    async fn find_all_by_livestream_id_limit(
        &self,
        conn: &mut DBConn,
        livestream_id: i64,
        limit: i64,
    ) -> isupipe_core::repos::Result<Vec<ReactionModel>> {
        let reaction_models: Vec<ReactionModel> = sqlx::query_as(
            "SELECT * FROM reactions WHERE livestream_id = ? ORDER BY created_at DESC LIMIT ?",
        )
        .bind(livestream_id)
        .bind(limit)
        .fetch_all(conn)
        .await?;

        Ok(reaction_models)
    }
}
