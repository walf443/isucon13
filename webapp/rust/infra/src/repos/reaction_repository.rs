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
#[cfg(test)]
mod most_favorite_emoji_by_livestream_user_name;

use crate::sql_support::{optional_scalar_string, scalar_i64};
use crate::tables::reaction::ReactionRow;
use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::reaction::{CreateReaction, Reaction, ReactionId};
use isupipe_core::models::user::{UserId, UserName};
use isupipe_core::repos::reaction_repository::ReactionRepository;

#[derive(Clone)]
pub struct ReactionRepositoryInfra {}

#[async_trait]
impl ReactionRepository for ReactionRepositoryInfra {
    async fn create<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        reaction: &CreateReaction,
    ) -> isupipe_core::repos::Result<ReactionId> {
        let row = ReactionRow::create()
            .user_id(&reaction.user_id)
            .livestream_id(&reaction.livestream_id)
            .emoji_name(&reaction.emoji_name)
            .created_at(reaction.created_at)
            .exec(conn)
            .await?;

        Ok(row.id)
    }

    async fn count_by_livestream_id<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        livestream_id: &LivestreamId,
    ) -> isupipe_core::repos::Result<i64> {
        let rows = toasty::sql::query(
            r#"
            SELECT COUNT(*)
            FROM livestreams l
            INNER JOIN reactions r ON l.id = r.livestream_id
            WHERE l.id = ?
            "#,
        )
        .bind(*livestream_id.inner())
        .exec(conn)
        .await?;

        Ok(scalar_i64(rows)?)
    }

    async fn most_favorite_emoji_by_livestream_user_name<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        livestream_user_name: &UserName,
    ) -> isupipe_core::repos::Result<String> {
        let rows = toasty::sql::query(
            r#"
            SELECT r.emoji_name
            FROM users u
            INNER JOIN livestreams l ON l.user_id = u.id
            INNER JOIN reactions r ON r.livestream_id = l.id
            WHERE u.name = ?
            GROUP BY emoji_name
            ORDER BY COUNT(*) DESC, emoji_name DESC
            LIMIT 1
            "#,
        )
        .bind(livestream_user_name.inner().as_str())
        .exec(conn)
        .await?;

        Ok(optional_scalar_string(rows)?.unwrap_or_default())
    }

    async fn count_by_livestream_user_id<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        livestream_user_id: &UserId,
    ) -> isupipe_core::repos::Result<i64> {
        let rows = toasty::sql::query(
            r#"
            SELECT COUNT(*)
            FROM users u
            INNER JOIN livestreams l ON l.user_id = u.id
            INNER JOIN reactions r ON r.livestream_id = l.id
            WHERE u.id = ?
            "#,
        )
        .bind(*livestream_user_id.inner())
        .exec(conn)
        .await?;

        Ok(scalar_i64(rows)?)
    }

    async fn count_by_livestream_user_name<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        livestream_user_name: &UserName,
    ) -> isupipe_core::repos::Result<i64> {
        let rows = toasty::sql::query(
            r#"
            SELECT COUNT(*)
            FROM users u
            INNER JOIN livestreams l ON l.user_id = u.id
            INNER JOIN reactions r ON r.livestream_id = l.id
            WHERE u.name = ?
            "#,
        )
        .bind(livestream_user_name.inner().as_str())
        .exec(conn)
        .await?;

        Ok(scalar_i64(rows)?)
    }

    async fn find_all_by_livestream_id<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        livestream_id: &LivestreamId,
    ) -> isupipe_core::repos::Result<Vec<Reaction>> {
        let rows = ReactionRow::filter(ReactionRow::fields().livestream_id().eq(livestream_id))
            .order_by(ReactionRow::fields().created_at().desc())
            .exec(conn)
            .await?;

        Ok(rows.into_iter().map(Into::into).collect())
    }

    async fn find_all_by_livestream_id_limit<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        livestream_id: &LivestreamId,
        limit: i64,
    ) -> isupipe_core::repos::Result<Vec<Reaction>> {
        let rows = ReactionRow::filter(ReactionRow::fields().livestream_id().eq(livestream_id))
            .order_by(ReactionRow::fields().created_at().desc())
            .limit(limit as usize)
            .exec(conn)
            .await?;

        Ok(rows.into_iter().map(Into::into).collect())
    }
}
