use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::reaction::ReactionModel;
use isupipe_core::repos::reaction_repository::ReactionRepository;

pub struct ReactionRepositoryInfra {}

#[async_trait]
impl ReactionRepository for ReactionRepositoryInfra {
    async fn insert(
        &self,
        conn: &mut DBConn,
        user_id: i64,
        livestream_id: i64,
        emoji_name: &str,
        created_at: i64,
    ) -> isupipe_core::repos::Result<i64> {
        let result =
            sqlx::query("INSERT INTO reactions (user_id, livestream_id, emoji_name, created_at) VALUES (?, ?, ?, ?)")
                .bind(user_id)
                .bind(livestream_id)
                .bind(emoji_name)
                .bind(created_at)
                .execute(conn)
                .await?;
        let reaction_id = result.last_insert_id() as i64;

        Ok(reaction_id)
    }

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
