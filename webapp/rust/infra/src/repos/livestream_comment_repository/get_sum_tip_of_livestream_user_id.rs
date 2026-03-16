use crate::repos::livestream_comment_repository::LivestreamCommentRepositoryInfra;
use isupipe_core::db::get_db_pool;
use isupipe_core::models::user::UserId;
use isupipe_core::repos::livestream_comment_repository::LivestreamCommentRepository;
use crate::qbey_support::bind_qbey_values;
use crate::tables::livestream::TABLE_LIVESTREAMS;
use crate::tables::livestream_comment::TABLE_LIVECOMMENTS;
use crate::tables::user::TABLE_USERS;
use qbey_mysql::qbey;

#[tokio::test]
async fn empty_case() {
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
        ins.add_value(&[("id", 1i64.into()), ("user_id", 1i64.into()), ("title", "t".into()), ("description", "d".into()), ("playlist_url", "http://p".into()), ("thumbnail_url", "http://t".into()), ("start_at", 1000i64.into()), ("end_at", 2000i64.into())]);
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds).execute(&mut *tx).await.unwrap();
    }

    let repo = LivestreamCommentRepositoryInfra {};
    let total = repo
        .get_sum_tip_of_livestream_user_id(&mut tx, &UserId::new(1))
        .await
        .unwrap();
    assert_eq!(total, 0);
}

#[tokio::test]
async fn not_empty_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    {
        let mut ins = qbey(TABLE_USERS.table()).into_insert();
        ins.add_value(&[("id", 1i64.into()), ("name", "alice".into()), ("display_name", "Alice".into()), ("password", "pw".into()), ("description", "desc".into())]);
        ins.add_value(&[("id", 2i64.into()), ("name", "bob".into()), ("display_name", "Bob".into()), ("password", "pw".into()), ("description", "desc".into())]);
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds).execute(&mut *tx).await.unwrap();
    }

    {
        let mut ins = qbey(TABLE_LIVESTREAMS.table()).into_insert();
        ins.add_value(&[("id", 1i64.into()), ("user_id", 1i64.into()), ("title", "t".into()), ("description", "d".into()), ("playlist_url", "http://p".into()), ("thumbnail_url", "http://t".into()), ("start_at", 1000i64.into()), ("end_at", 2000i64.into())]);
        ins.add_value(&[("id", 2i64.into()), ("user_id", 1i64.into()), ("title", "t2".into()), ("description", "d2".into()), ("playlist_url", "http://p2".into()), ("thumbnail_url", "http://t2".into()), ("start_at", 3000i64.into()), ("end_at", 4000i64.into())]);
        ins.add_value(&[("id", 3i64.into()), ("user_id", 2i64.into()), ("title", "t3".into()), ("description", "d3".into()), ("playlist_url", "http://p3".into()), ("thumbnail_url", "http://t3".into()), ("start_at", 5000i64.into()), ("end_at", 6000i64.into())]);
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds).execute(&mut *tx).await.unwrap();
    }

    {
        let mut ins = qbey(TABLE_LIVECOMMENTS.table()).into_insert();
        ins.add_value(&[("user_id", 2i64.into()), ("livestream_id", 1i64.into()), ("comment", "a".into()), ("tip", 100i64.into()), ("created_at", 1i64.into())]);
        ins.add_value(&[("user_id", 2i64.into()), ("livestream_id", 1i64.into()), ("comment", "b".into()), ("tip", 200i64.into()), ("created_at", 2i64.into())]);
        ins.add_value(&[("user_id", 2i64.into()), ("livestream_id", 2i64.into()), ("comment", "c".into()), ("tip", 50i64.into()), ("created_at", 3i64.into())]);
        ins.add_value(&[("user_id", 1i64.into()), ("livestream_id", 3i64.into()), ("comment", "d".into()), ("tip", 300i64.into()), ("created_at", 4i64.into())]);
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds).execute(&mut *tx).await.unwrap();
    }

    let repo = LivestreamCommentRepositoryInfra {};

    // alice owns livestream 1 and 2, tips: 100+200+50=350
    let total = repo
        .get_sum_tip_of_livestream_user_id(&mut tx, &UserId::new(1))
        .await
        .unwrap();
    assert_eq!(total, 350);

    // bob owns livestream 3, tips: 300
    let total = repo
        .get_sum_tip_of_livestream_user_id(&mut tx, &UserId::new(2))
        .await
        .unwrap();
    assert_eq!(total, 300);
}
