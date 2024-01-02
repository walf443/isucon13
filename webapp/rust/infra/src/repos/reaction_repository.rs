use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::mysql_decimal::MysqlDecimal;
use isupipe_core::models::reaction::{CreateReaction, Reaction, ReactionId};
use isupipe_core::models::user::UserId;
use isupipe_core::repos::reaction_repository::ReactionRepository;

#[derive(Clone)]
pub struct ReactionRepositoryInfra {}

#[async_trait]
impl ReactionRepository for ReactionRepositoryInfra {
    async fn create(
        &self,
        conn: &mut DBConn,
        reaction: &CreateReaction,
    ) -> isupipe_core::repos::Result<ReactionId> {
        let result =
            sqlx::query("INSERT INTO reactions (user_id, livestream_id, emoji_name, created_at) VALUES (?, ?, ?, ?)")
                .bind(&reaction.user_id)
                .bind(&reaction.livestream_id)
                .bind(&reaction.emoji_name)
                .bind(reaction.created_at)
                .execute(conn)
                .await?;
        let reaction_id = result.last_insert_id() as i64;

        Ok(ReactionId::new(reaction_id))
    }

    async fn count_by_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
    ) -> isupipe_core::repos::Result<i64> {
        let MysqlDecimal(reactions) = sqlx::query_scalar("SELECT COUNT(*) FROM livestreams l INNER JOIN reactions r ON l.id = r.livestream_id WHERE l.id = ?")
            .bind(livestream_id)
            .fetch_one(conn)
            .await?;

        Ok(reactions)
    }

    async fn most_favorite_emoji_by_livestream_user_name(
        &self,
        conn: &mut DBConn,
        livestream_user_name: &str,
    ) -> isupipe_core::repos::Result<String> {
        let query = r#"
            SELECT r.emoji_name
            FROM users u
            INNER JOIN livestreams l ON l.user_id = u.id
            INNER JOIN reactions r ON r.livestream_id = l.id
            WHERE u.name = ?
            GROUP BY emoji_name
            ORDER BY COUNT(*) DESC, emoji_name DESC
            LIMIT 1
        "#;
        let favorite_emoji: String = sqlx::query_scalar(query)
            .bind(livestream_user_name)
            .fetch_optional(conn)
            .await?
            .unwrap_or_default();

        Ok(favorite_emoji)
    }

    async fn count_by_livestream_user_id(
        &self,
        conn: &mut DBConn,
        livestream_user_id: &UserId,
    ) -> isupipe_core::repos::Result<i64> {
        let query = r#"
            SELECT COUNT(*) FROM users u
            INNER JOIN livestreams l ON l.user_id = u.id
            INNER JOIN reactions r ON r.livestream_id = l.id
            WHERE u.id = ?
        "#;
        let MysqlDecimal(reactions) = sqlx::query_scalar(query)
            .bind(livestream_user_id)
            .fetch_one(conn)
            .await?;

        Ok(reactions)
    }

    async fn count_by_livestream_user_name(
        &self,
        conn: &mut DBConn,
        livestream_user_name: &str,
    ) -> isupipe_core::repos::Result<i64> {
        let query = r"#
            SELECT COUNT(*) FROM users u
            INNER JOIN livestreams l ON l.user_id = u.id
            INNER JOIN reactions r ON r.livestream_id = l.id
            WHERE u.name = ?
        #";
        let MysqlDecimal(total_reactions) = sqlx::query_scalar(query)
            .bind(livestream_user_name)
            .fetch_one(conn)
            .await?;

        Ok(total_reactions)
    }

    async fn find_all_by_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
    ) -> isupipe_core::repos::Result<Vec<Reaction>> {
        let reaction_models: Vec<Reaction> = sqlx::query_as(
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
        livestream_id: &LivestreamId,
        limit: i64,
    ) -> isupipe_core::repos::Result<Vec<Reaction>> {
        let reaction_models: Vec<Reaction> = sqlx::query_as(
            "SELECT * FROM reactions WHERE livestream_id = ? ORDER BY created_at DESC LIMIT ?",
        )
        .bind(livestream_id)
        .bind(limit)
        .fetch_all(conn)
        .await?;

        Ok(reaction_models)
    }
}
