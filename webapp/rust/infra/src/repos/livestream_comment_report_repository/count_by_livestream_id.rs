use crate::qbey_support::bind_qbey_values;
use crate::repos::livestream_comment_report_repository::LivestreamCommentReportRepositoryInfra;
use crate::tables::livestream::TABLE_LIVESTREAMS;
use crate::tables::livestream_comment::TABLE_LIVECOMMENTS;
use crate::tables::livestream_comment_report::TABLE_LIVECOMMENT_REPORTS;
use crate::tables::user::TABLE_USERS;
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::repos::livestream_comment_report_repository::LivestreamCommentReportRepository;
use qbey_mysql::qbey;

#[tokio::test]
async fn zero_case() {
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

    let repo = LivestreamCommentReportRepositoryInfra {};
    let result = repo
        .count_by_livestream_id(&mut tx, &LivestreamId::new(1))
        .await
        .unwrap();
    assert_eq!(result, 0);
}

#[tokio::test]
async fn counts_correctly() {
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
        ins.add_value(&[("id", 2i64.into()), ("user_id", 1i64.into()), ("title", "t2".into()), ("description", "d2".into()), ("playlist_url", "http://p2".into()), ("thumbnail_url", "http://t2".into()), ("start_at", 3000i64.into()), ("end_at", 4000i64.into())]);
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds).execute(&mut *tx).await.unwrap();
    }

    {
        let mut ins = qbey(TABLE_LIVECOMMENTS.table()).into_insert();
        ins.add_value(&[("id", 1i64.into()), ("user_id", 1i64.into()), ("livestream_id", 1i64.into()), ("comment", "a".into()), ("tip", 0i64.into()), ("created_at", 100i64.into())]);
        ins.add_value(&[("id", 2i64.into()), ("user_id", 1i64.into()), ("livestream_id", 1i64.into()), ("comment", "b".into()), ("tip", 0i64.into()), ("created_at", 200i64.into())]);
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds).execute(&mut *tx).await.unwrap();
    }

    {
        let mut ins = qbey(TABLE_LIVECOMMENT_REPORTS.table()).into_insert();
        ins.add_value(&[("id", 1i64.into()), ("user_id", 1i64.into()), ("livestream_id", 1i64.into()), ("livecomment_id", 1i64.into()), ("created_at", 100i64.into())]);
        ins.add_value(&[("id", 2i64.into()), ("user_id", 1i64.into()), ("livestream_id", 1i64.into()), ("livecomment_id", 2i64.into()), ("created_at", 200i64.into())]);
        ins.add_value(&[("id", 3i64.into()), ("user_id", 1i64.into()), ("livestream_id", 2i64.into()), ("livecomment_id", 1i64.into()), ("created_at", 300i64.into())]);
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds).execute(&mut *tx).await.unwrap();
    }

    let repo = LivestreamCommentReportRepositoryInfra {};
    let result = repo
        .count_by_livestream_id(&mut tx, &LivestreamId::new(1))
        .await
        .unwrap();
    assert_eq!(result, 2);

    let result = repo
        .count_by_livestream_id(&mut tx, &LivestreamId::new(2))
        .await
        .unwrap();
    assert_eq!(result, 1);
}
