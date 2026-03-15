#[cfg(test)]
mod find;
#[cfg(test)]
mod find_all;
#[cfg(test)]
mod find_all_by_livestream_id;
#[cfg(test)]
mod find_all_by_livestream_id_order_by_created_at;
#[cfg(test)]
mod find_all_by_livestream_id_order_by_created_at_limit;
#[cfg(test)]
mod get_max_tip_of_livestream_id;
#[cfg(test)]
mod get_sum_tip;
#[cfg(test)]
mod get_sum_tip_of_livestream_id;
#[cfg(test)]
mod get_sum_tip_of_livestream_user_id;

use crate::qbey_support::bind_qbey_values;
use crate::tables::livestream::TABLE_LIVESTREAMS;
use crate::tables::livestream_comment::TABLE_LIVECOMMENTS;
use crate::tables::user::TABLE_USERS;
use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::livestream_comment::{
    CreateLivestreamComment, LivestreamComment, LivestreamCommentId,
};
use isupipe_core::models::user::UserId;
use isupipe_core::repos::livestream_comment_repository::LivestreamCommentRepository;
use qbey::RawSql;
use qbey_mysql::qbey;

#[derive(Clone)]
pub struct LivestreamCommentRepositoryInfra {}

#[async_trait]
impl LivestreamCommentRepository for LivestreamCommentRepositoryInfra {
    async fn create(
        &self,
        conn: &mut DBConn,
        comment: &CreateLivestreamComment,
    ) -> isupipe_core::repos::Result<LivestreamCommentId> {
        let rs = sqlx::query(
            "INSERT INTO livecomments (user_id, livestream_id, comment, tip, created_at) VALUES (?, ?, ?, ?, ?)",
        )
            .bind(&comment.user_id)
            .bind(&comment.livestream_id)
            .bind(&comment.comment)
            .bind(comment.tip)
            .bind(comment.created_at)
            .execute(conn)
            .await?;
        let comment_id = rs.last_insert_id() as i64;

        Ok(LivestreamCommentId::new(comment_id))
    }

    async fn remove_if_match_ng_word(
        &self,
        conn: &mut DBConn,
        comment: &LivestreamComment,
        ng_word: &str,
    ) -> isupipe_core::repos::Result<()> {
        // DELETE with complex subquery - not supported by qbey
        let query = r#"
        DELETE FROM livecomments
        WHERE
        id = ? AND
        livestream_id = ? AND
        (SELECT COUNT(*)
        FROM
        (SELECT ? AS text) AS texts
        INNER JOIN
        (SELECT CONCAT('%', ?, '%')	AS pattern) AS patterns
        ON texts.text LIKE patterns.pattern) >= 1
        "#;
        sqlx::query(query)
            .bind(&comment.id)
            .bind(&comment.livestream_id)
            .bind(&comment.comment)
            .bind(ng_word)
            .execute(conn)
            .await?;

        Ok(())
    }

    async fn find(
        &self,
        conn: &mut DBConn,
        comment_id: &LivestreamCommentId,
    ) -> isupipe_core::repos::Result<Option<LivestreamComment>> {
        let t = &TABLE_LIVECOMMENTS;
        let mut q = qbey(t.table_name());
        q.and_where(t.id().eq(*comment_id.inner()));
        let (sql, binds) = q.to_sql();
        let comment = bind_qbey_values!(sqlx::query_as::<_, LivestreamComment>(&sql), binds)
            .fetch_optional(conn)
            .await?;

        Ok(comment)
    }

    async fn find_all(
        &self,
        conn: &mut DBConn,
    ) -> isupipe_core::repos::Result<Vec<LivestreamComment>> {
        let t = &TABLE_LIVECOMMENTS;
        let q = qbey(t.table_name());
        let (sql, binds) = q.to_sql();
        let livecomments =
            bind_qbey_values!(sqlx::query_as::<_, LivestreamComment>(&sql), binds)
                .fetch_all(conn)
                .await?;

        Ok(livecomments)
    }

    async fn find_all_by_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
    ) -> isupipe_core::repos::Result<Vec<LivestreamComment>> {
        let t = &TABLE_LIVECOMMENTS;
        let mut q = qbey(t.table_name());
        q.and_where(t.livestream_id().eq(*livestream_id.inner()));
        let (sql, binds) = q.to_sql();
        let comments =
            bind_qbey_values!(sqlx::query_as::<_, LivestreamComment>(&sql), binds)
                .fetch_all(conn)
                .await?;

        Ok(comments)
    }

    async fn find_all_by_livestream_id_order_by_created_at(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
    ) -> isupipe_core::repos::Result<Vec<LivestreamComment>> {
        let t = &TABLE_LIVECOMMENTS;
        let mut q = qbey(t.table_name());
        q.and_where(t.livestream_id().eq(*livestream_id.inner()));
        q.order_by(t.created_at().desc());
        let (sql, binds) = q.to_sql();
        let comments =
            bind_qbey_values!(sqlx::query_as::<_, LivestreamComment>(&sql), binds)
                .fetch_all(conn)
                .await?;

        Ok(comments)
    }

    async fn find_all_by_livestream_id_order_by_created_at_limit(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
        limit: i64,
    ) -> isupipe_core::repos::Result<Vec<LivestreamComment>> {
        let t = &TABLE_LIVECOMMENTS;
        let mut q = qbey(t.table_name());
        q.and_where(t.livestream_id().eq(*livestream_id.inner()));
        q.order_by(t.created_at().desc());
        q.limit(limit as u64);
        let (sql, binds) = q.to_sql();
        let comments =
            bind_qbey_values!(sqlx::query_as::<_, LivestreamComment>(&sql), binds)
                .fetch_all(conn)
                .await?;

        Ok(comments)
    }

    async fn get_sum_tip(&self, conn: &mut DBConn) -> isupipe_core::repos::Result<i64> {
        let t = &TABLE_LIVECOMMENTS;
        let mut q = qbey(t.table_name());
        q.add_select_expr(RawSql::new("CAST(IFNULL(SUM(tip), 0) AS SIGNED)"), None);
        let (sql, binds) = q.to_sql();
        let total_tip = bind_qbey_values!(sqlx::query_scalar::<_, i64>(&sql), binds)
            .fetch_one(conn)
            .await?;

        Ok(total_tip)
    }

    async fn get_sum_tip_of_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
    ) -> isupipe_core::repos::Result<i64> {
        let livestream = &TABLE_LIVESTREAMS;
        let comment = &TABLE_LIVECOMMENTS;
        let mut q = qbey(livestream.table_name());
        q.as_("l");
        let l = livestream.as_("l");
        q.join(
            comment.table_name(),
            l.id().eq_col(comment.livestream_id()),
        );
        q.and_where(l.id().eq(*livestream_id.inner()));
        q.add_select_expr(RawSql::new("CAST(IFNULL(SUM(livecomments.tip), 0) AS SIGNED)"), None);
        let (sql, binds) = q.to_sql();
        let total_tips = bind_qbey_values!(sqlx::query_scalar::<_, i64>(&sql), binds)
            .fetch_one(conn)
            .await?;

        Ok(total_tips)
    }

    async fn get_max_tip_of_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
    ) -> isupipe_core::repos::Result<i64> {
        let livestream = &TABLE_LIVESTREAMS;
        let comment = &TABLE_LIVECOMMENTS;
        let mut q = qbey(livestream.table_name());
        q.as_("l");
        let l = livestream.as_("l");
        q.join(
            comment.table_name(),
            l.id().eq_col(comment.livestream_id()),
        );
        q.and_where(l.id().eq(*livestream_id.inner()));
        q.add_select_expr(RawSql::new("CAST(IFNULL(MAX(livecomments.tip), 0) AS SIGNED)"), None);

        let (sql, binds) = q.to_sql();
        let max_tip = bind_qbey_values!(sqlx::query_scalar::<_, i64>(&sql), binds)
            .fetch_one(conn)
            .await?;

        Ok(max_tip)
    }

    async fn get_sum_tip_of_livestream_user_id(
        &self,
        conn: &mut DBConn,
        user_id: &UserId,
    ) -> isupipe_core::repos::Result<i64> {
        let user = &TABLE_USERS;
        let livestream = &TABLE_LIVESTREAMS;
        let comment = &TABLE_LIVECOMMENTS;
        let mut q = qbey(user.table_name());
        q.as_("u");
        let u = user.as_("u");
        q.join(
            livestream.table_name(),
            u.id().eq_col(livestream.user_id()),
        );
        q.join(
            comment.table_name(),
            livestream.id().eq_col(comment.livestream_id()),
        );
        q.and_where(u.id().eq(*user_id.inner()));
        q.add_select_expr(RawSql::new("CAST(IFNULL(SUM(livecomments.tip), 0) AS SIGNED)"), None);

        let (sql, binds) = q.to_sql();
        let tips = bind_qbey_values!(sqlx::query_scalar::<_, i64>(&sql), binds)
            .fetch_one(conn)
            .await?;

        Ok(tips)
    }
}
