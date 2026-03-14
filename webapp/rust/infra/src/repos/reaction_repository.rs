#[cfg(test)]
mod count_by_livestream_id;
#[cfg(test)]
mod count_by_livestream_user_id;
#[cfg(test)]
mod count_by_livestream_user_name;
#[cfg(test)]
mod create;
#[cfg(test)]
mod find_all_by_livestream_id;
#[cfg(test)]
mod find_all_by_livestream_id_limit;

use crate::sqipe_support::bind_sqipe_values;
use crate::tables::livestream::TABLE_LIVESTREAMS;
use crate::tables::reaction::TABLE_REACTIONS;
use crate::tables::user::TABLE_USERS;
use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::reaction::{CreateReaction, Reaction, ReactionId};
use isupipe_core::models::user::{UserId, UserName};
use isupipe_core::repos::reaction_repository::ReactionRepository;
use sqipe::{aggregate, table};
use sqipe_mysql::sqipe;

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
        let livestream = &TABLE_LIVESTREAMS;
        let reaction = &TABLE_REACTIONS;
        let mut q = sqipe(livestream.table_name());
        q.as_("l");
        q.join(reaction.table_name(), table("l").col("id").eq_col("livestream_id"));
        q.and_where(table("l").col("id").eq(*livestream_id.inner()));
        q.aggregate(&[aggregate::count_all()]);
        let (sql, binds) = q.to_sql();
        let reactions = bind_sqipe_values!(sqlx::query_scalar::<_, i64>(&sql), binds)
            .fetch_one(conn)
            .await?;

        Ok(reactions)
    }

    async fn most_favorite_emoji_by_livestream_user_name(
        &self,
        conn: &mut DBConn,
        livestream_user_name: &UserName,
    ) -> isupipe_core::repos::Result<String> {
        // ORDER BY COUNT(*) DESC - not supported by squipe
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
        let user = &TABLE_USERS;
        let livestream = &TABLE_LIVESTREAMS;
        let reaction = &TABLE_REACTIONS;
        let mut q = sqipe(user.table_name());
        q.as_("u");
        q.join(livestream.table_name(), table("u").col("id").eq_col("user_id"));
        q.join(reaction.table_name(), livestream.id().eq_col(reaction.livestream_id()));
        q.and_where(table("u").col("id").eq(*livestream_user_id.inner()));
        q.aggregate(&[aggregate::count_all()]);
        let (sql, binds) = q.to_sql();
        let reactions = bind_sqipe_values!(sqlx::query_scalar::<_, i64>(&sql), binds)
            .fetch_one(conn)
            .await?;

        Ok(reactions)
    }

    async fn count_by_livestream_user_name(
        &self,
        conn: &mut DBConn,
        livestream_user_name: &UserName,
    ) -> isupipe_core::repos::Result<i64> {
        let user = &TABLE_USERS;
        let livestream = &TABLE_LIVESTREAMS;
        let reaction = &TABLE_REACTIONS;
        let mut q = sqipe(user.table_name());
        q.as_("u");
        q.join(livestream.table_name(), table("u").col("id").eq_col("user_id"));
        q.join(reaction.table_name(), livestream.id().eq_col(reaction.livestream_id()));
        q.and_where(table("u").col("name").eq(livestream_user_name.inner().clone()));
        q.aggregate(&[aggregate::count_all()]);
        let (sql, binds) = q.to_sql();
        let total_reactions = bind_sqipe_values!(sqlx::query_scalar::<_, i64>(&sql), binds)
            .fetch_one(conn)
            .await?;

        Ok(total_reactions)
    }

    async fn find_all_by_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
    ) -> isupipe_core::repos::Result<Vec<Reaction>> {
        let t = &TABLE_REACTIONS;
        let mut q = sqipe(t.table_name());
        q.and_where(t.livestream_id().eq(*livestream_id.inner()));
        q.order_by(t.created_at().desc());
        let (sql, binds) = q.to_sql();
        let reaction_models = bind_sqipe_values!(sqlx::query_as::<_, Reaction>(&sql), binds)
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
        let t = &TABLE_REACTIONS;
        let mut q = sqipe(t.table_name());
        q.and_where(t.livestream_id().eq(*livestream_id.inner()));
        q.order_by(t.created_at().desc());
        q.limit(limit as u64);
        let (sql, binds) = q.to_sql();
        let reaction_models = bind_sqipe_values!(sqlx::query_as::<_, Reaction>(&sql), binds)
            .fetch_all(conn)
            .await?;

        Ok(reaction_models)
    }
}
