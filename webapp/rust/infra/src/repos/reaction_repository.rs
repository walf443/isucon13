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

use crate::qbey_support::bind_qbey_values;
use crate::tables::livestream::TABLE_LIVESTREAMS;
use crate::tables::reaction::TABLE_REACTIONS;
use crate::tables::user::TABLE_USERS;
use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::reaction::{CreateReaction, Reaction, ReactionId};
use isupipe_core::models::user::{UserId, UserName};
use isupipe_core::repos::reaction_repository::ReactionRepository;
use qbey::prelude::*;
use qbey_mysql::qbey;

struct InsertReaction<'a>(&'a CreateReaction);

impl qbey::ToInsertRow<qbey::Value> for InsertReaction<'_> {
    fn to_insert_row(&self) -> Vec<(&'static str, qbey::Value)> {
        vec![
            ("user_id", (*self.0.user_id.inner()).into()),
            ("livestream_id", (*self.0.livestream_id.inner()).into()),
            ("emoji_name", self.0.emoji_name.as_str().into()),
            ("created_at", self.0.created_at.into()),
        ]
    }
}

#[derive(Clone)]
pub struct ReactionRepositoryInfra {}

#[async_trait]
impl ReactionRepository for ReactionRepositoryInfra {
    async fn create(
        &self,
        conn: &mut DBConn,
        reaction: &CreateReaction,
    ) -> isupipe_core::repos::Result<ReactionId> {
        let t = &TABLE_REACTIONS;
        let mut ins = qbey(t.table()).into_insert();
        ins.add_value(&InsertReaction(reaction));
        let (sql, binds) = ins.into_sql();
        let result = bind_qbey_values!(sqlx::query(&sql), binds)
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
        let mut q = qbey(livestream.table());
        q.as_("l");
        let l = livestream.as_("l");
        q.join(reaction.table(), l.id().eq(reaction.livestream_id()));
        q.and_where(l.id().eq(*livestream_id.inner()));
        q.add_select(qbey::count_all());
        let (sql, binds) = q.into_sql();
        let reactions = bind_qbey_values!(sqlx::query_scalar::<_, i64>(&sql), binds)
            .fetch_one(conn)
            .await?;

        Ok(reactions)
    }

    async fn most_favorite_emoji_by_livestream_user_name(
        &self,
        conn: &mut DBConn,
        livestream_user_name: &UserName,
    ) -> isupipe_core::repos::Result<String> {
        let user = &TABLE_USERS;
        let livestream = &TABLE_LIVESTREAMS;
        let reaction = &TABLE_REACTIONS;
        let mut q = qbey(user.table());
        q.as_("u");
        let u = user.as_("u");
        q.join(livestream.table(), u.id().eq(livestream.user_id()));
        q.join(
            reaction.table(),
            livestream.id().eq(reaction.livestream_id()),
        );
        q.and_where(u.name().eq(livestream_user_name.inner().clone()));
        q.select(&[reaction.emoji_name().into(), qbey::count_all().as_("cnt")]);
        q.group_by(&["emoji_name"]);
        q.order_by(qbey::col("cnt").desc());
        q.order_by(qbey::col("emoji_name").desc());
        q.limit(1);
        let (sql, binds) = q.into_sql();
        let favorite_emoji: String =
            bind_qbey_values!(sqlx::query_scalar::<_, String>(&sql), binds)
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
        let mut q = qbey(user.table());
        q.as_("u");
        let u = user.as_("u");
        q.join(livestream.table(), u.id().eq(livestream.user_id()));
        q.join(
            reaction.table(),
            livestream.id().eq(reaction.livestream_id()),
        );
        q.and_where(u.id().eq(*livestream_user_id.inner()));
        q.add_select(qbey::count_all());
        let (sql, binds) = q.into_sql();
        let reactions = bind_qbey_values!(sqlx::query_scalar::<_, i64>(&sql), binds)
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
        let mut q = qbey(user.table());
        q.as_("u");
        let u = user.as_("u");
        q.join(livestream.table(), u.id().eq(livestream.user_id()));
        q.join(
            reaction.table(),
            livestream.id().eq(reaction.livestream_id()),
        );
        q.and_where(u.name().eq(livestream_user_name.inner().clone()));
        q.add_select(qbey::count_all());
        let (sql, binds) = q.into_sql();
        let total_reactions = bind_qbey_values!(sqlx::query_scalar::<_, i64>(&sql), binds)
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
        let mut q = qbey(t.table());
        q.and_where(t.livestream_id().eq(*livestream_id.inner()));
        q.order_by(t.created_at().desc());
        let (sql, binds) = q.into_sql();
        let reaction_models = bind_qbey_values!(sqlx::query_as::<_, Reaction>(&sql), binds)
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
        let mut q = qbey(t.table());
        q.and_where(t.livestream_id().eq(*livestream_id.inner()));
        q.order_by(t.created_at().desc());
        q.limit(limit as u64);
        let (sql, binds) = q.into_sql();
        let reaction_models = bind_qbey_values!(sqlx::query_as::<_, Reaction>(&sql), binds)
            .fetch_all(conn)
            .await?;

        Ok(reaction_models)
    }
}
