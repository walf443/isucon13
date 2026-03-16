use crate::repos::livestream_comment_repository::LivestreamCommentRepositoryInfra;
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::repos::livestream_comment_repository::LivestreamCommentRepository;
use crate::qbey_support::bind_qbey_values;
use crate::tables::livestream::TABLE_LIVESTREAMS;
use crate::tables::livestream_comment::TABLE_LIVECOMMENTS;
use crate::tables::user::TABLE_USERS;
use qbey_mysql::qbey;

#[tokio::test]
async fn returns_ordered_by_created_at_desc() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    {
        let mut ins = qbey(TABLE_USERS.table()).into_insert();
        ins.add_value(&[("id", 1i64.into()), ("name", "test".into()), ("display_name", "Test".into()), ("password", "pw".into()), ("description", "desc".into())]);
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds).execute(&mut *tx).await.unwrap();
    }

    {
        let mut ins = qbey(TABLE_LIVESTREAMS.table()).into_insert();
        ins.add_value(&[("id", 1i64.into()), ("user_id", 1i64.into()), ("title", "t1".into()), ("description", "d1".into()), ("playlist_url", "http://p".into()), ("thumbnail_url", "http://t".into()), ("start_at", 1000i64.into()), ("end_at", 2000i64.into())]);
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds).execute(&mut *tx).await.unwrap();
    }

    {
        let mut ins = qbey(TABLE_LIVECOMMENTS.table()).into_insert();
        ins.add_value(&[("id", 1i64.into()), ("user_id", 1i64.into()), ("livestream_id", 1i64.into()), ("comment", "a".into()), ("tip", 0i64.into()), ("created_at", 100i64.into())]);
        ins.add_value(&[("id", 2i64.into()), ("user_id", 1i64.into()), ("livestream_id", 1i64.into()), ("comment", "b".into()), ("tip", 0i64.into()), ("created_at", 300i64.into())]);
        ins.add_value(&[("id", 3i64.into()), ("user_id", 1i64.into()), ("livestream_id", 1i64.into()), ("comment", "c".into()), ("tip", 0i64.into()), ("created_at", 200i64.into())]);
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds).execute(&mut *tx).await.unwrap();
    }

    let repo = LivestreamCommentRepositoryInfra {};
    let result = repo
        .find_all_by_livestream_id_order_by_created_at(&mut tx, &LivestreamId::new(1))
        .await
        .unwrap();
    assert_eq!(result.len(), 3);
    assert_eq!(result[0].created_at, 300);
    assert_eq!(result[1].created_at, 200);
    assert_eq!(result[2].created_at, 100);
}
