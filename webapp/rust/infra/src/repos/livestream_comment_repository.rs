#[cfg(test)]
mod create;
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

use crate::sql_support::scalar_i64;
use crate::tables::livestream_comment::LivestreamCommentRow;
use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::livestream_comment::{
    CreateLivestreamComment, LivestreamComment, LivestreamCommentId,
};
use isupipe_core::models::user::UserId;
use isupipe_core::repos::livestream_comment_repository::LivestreamCommentRepository;

#[derive(Clone)]
pub struct LivestreamCommentRepositoryInfra {}

#[async_trait]
impl LivestreamCommentRepository for LivestreamCommentRepositoryInfra {
    async fn create<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        comment: &CreateLivestreamComment,
    ) -> isupipe_core::repos::Result<LivestreamCommentId> {
        let row = LivestreamCommentRow::create()
            .user_id(&comment.user_id)
            .livestream_id(&comment.livestream_id)
            .comment(&comment.comment)
            .tip(comment.tip)
            .created_at(comment.created_at)
            .exec(conn)
            .await?;

        Ok(row.id)
    }

    async fn remove_if_match_ng_word<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        comment: &LivestreamComment,
        ng_word: &str,
    ) -> isupipe_core::repos::Result<()> {
        // DELETE with complex subquery - not supported by toasty's query builder
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
        toasty::sql::statement(query)
            .bind(*comment.id.inner())
            .bind(*comment.livestream_id.inner())
            .bind(comment.comment.as_str())
            .bind(ng_word)
            .exec(conn)
            .await?;

        Ok(())
    }

    async fn find<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        comment_id: &LivestreamCommentId,
    ) -> isupipe_core::repos::Result<Option<LivestreamComment>> {
        let row = LivestreamCommentRow::filter(LivestreamCommentRow::fields().id().eq(comment_id))
            .first()
            .exec(conn)
            .await?;

        Ok(row.map(Into::into))
    }

    async fn find_all<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
    ) -> isupipe_core::repos::Result<Vec<LivestreamComment>> {
        let rows = LivestreamCommentRow::all().exec(conn).await?;

        Ok(rows.into_iter().map(Into::into).collect())
    }

    async fn find_all_by_livestream_id<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        livestream_id: &LivestreamId,
    ) -> isupipe_core::repos::Result<Vec<LivestreamComment>> {
        let rows = LivestreamCommentRow::filter(
            LivestreamCommentRow::fields()
                .livestream_id()
                .eq(livestream_id),
        )
        .exec(conn)
        .await?;

        Ok(rows.into_iter().map(Into::into).collect())
    }

    async fn find_all_by_livestream_id_order_by_created_at<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        livestream_id: &LivestreamId,
    ) -> isupipe_core::repos::Result<Vec<LivestreamComment>> {
        let rows = LivestreamCommentRow::filter(
            LivestreamCommentRow::fields()
                .livestream_id()
                .eq(livestream_id),
        )
        .order_by(LivestreamCommentRow::fields().created_at().desc())
        .exec(conn)
        .await?;

        Ok(rows.into_iter().map(Into::into).collect())
    }

    async fn find_all_by_livestream_id_order_by_created_at_limit<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        livestream_id: &LivestreamId,
        limit: i64,
    ) -> isupipe_core::repos::Result<Vec<LivestreamComment>> {
        let rows = LivestreamCommentRow::filter(
            LivestreamCommentRow::fields()
                .livestream_id()
                .eq(livestream_id),
        )
        .order_by(LivestreamCommentRow::fields().created_at().desc())
        .limit(limit as usize)
        .exec(conn)
        .await?;

        Ok(rows.into_iter().map(Into::into).collect())
    }

    async fn get_sum_tip<'c>(&self, conn: &'c mut DBConn<'c>) -> isupipe_core::repos::Result<i64> {
        let rows =
            toasty::sql::query("SELECT CAST(IFNULL(SUM(tip), 0) AS SIGNED) FROM livecomments")
                .exec(conn)
                .await?;

        Ok(scalar_i64(rows)?)
    }

    async fn get_sum_tip_of_livestream_id<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        livestream_id: &LivestreamId,
    ) -> isupipe_core::repos::Result<i64> {
        let rows = toasty::sql::query(
            r#"
            SELECT CAST(IFNULL(SUM(l2.tip), 0) AS SIGNED)
            FROM livestreams l
            INNER JOIN livecomments l2 ON l.id = l2.livestream_id
            WHERE l.id = ?
            "#,
        )
        .bind(*livestream_id.inner())
        .exec(conn)
        .await?;

        Ok(scalar_i64(rows)?)
    }

    async fn get_max_tip_of_livestream_id<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        livestream_id: &LivestreamId,
    ) -> isupipe_core::repos::Result<i64> {
        let rows = toasty::sql::query(
            r#"
            SELECT CAST(IFNULL(MAX(l2.tip), 0) AS SIGNED)
            FROM livestreams l
            INNER JOIN livecomments l2 ON l.id = l2.livestream_id
            WHERE l.id = ?
            "#,
        )
        .bind(*livestream_id.inner())
        .exec(conn)
        .await?;

        Ok(scalar_i64(rows)?)
    }

    async fn get_sum_tip_of_livestream_user_id<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        user_id: &UserId,
    ) -> isupipe_core::repos::Result<i64> {
        let rows = toasty::sql::query(
            r#"
            SELECT CAST(IFNULL(SUM(l2.tip), 0) AS SIGNED)
            FROM users u
            INNER JOIN livestreams l ON l.user_id = u.id
            INNER JOIN livecomments l2 ON l2.livestream_id = l.id
            WHERE u.id = ?
            "#,
        )
        .bind(*user_id.inner())
        .exec(conn)
        .await?;

        Ok(scalar_i64(rows)?)
    }
}
