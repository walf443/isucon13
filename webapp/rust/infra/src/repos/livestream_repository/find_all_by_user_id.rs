use crate::qbey_support::bind_qbey_values;
use crate::repos::livestream_repository::LivestreamRepositoryInfra;
use crate::tables::livestream::TABLE_LIVESTREAMS;
use crate::tables::user::TABLE_USERS;
use isupipe_core::db::get_db_pool;
use isupipe_core::models::user::UserId;
use isupipe_core::repos::livestream_repository::LivestreamRepository;
use qbey_mysql::qbey;

#[tokio::test]
async fn empty_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    {
        let mut ins = qbey(TABLE_USERS.table()).into_insert();
        ins.add_value(&[
            ("id", 1i64.into()),
            ("name", "test".into()),
            ("display_name", "Test".into()),
            ("password", "pw".into()),
            ("description", "desc".into()),
        ]);
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let repo = LivestreamRepositoryInfra {};
    let result = repo
        .find_all_by_user_id(&mut tx, &UserId::new(1))
        .await
        .unwrap();
    assert_eq!(result.len(), 0);
}

#[tokio::test]
async fn filters_by_user_id() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    {
        let mut ins = qbey(TABLE_USERS.table()).into_insert();
        ins.add_value(&[
            ("id", 1i64.into()),
            ("name", "alice".into()),
            ("display_name", "Alice".into()),
            ("password", "pw".into()),
            ("description", "desc".into()),
        ]);
        ins.add_value(&[
            ("id", 2i64.into()),
            ("name", "bob".into()),
            ("display_name", "Bob".into()),
            ("password", "pw".into()),
            ("description", "desc".into()),
        ]);
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    {
        let mut ins = qbey(TABLE_LIVESTREAMS.table()).into_insert();
        ins.add_value(&[
            ("id", 1i64.into()),
            ("user_id", 1i64.into()),
            ("title", "t1".into()),
            ("description", "d1".into()),
            ("playlist_url", "http://p".into()),
            ("thumbnail_url", "http://t".into()),
            ("start_at", 1000i64.into()),
            ("end_at", 2000i64.into()),
        ]);
        ins.add_value(&[
            ("id", 2i64.into()),
            ("user_id", 1i64.into()),
            ("title", "t2".into()),
            ("description", "d2".into()),
            ("playlist_url", "http://p2".into()),
            ("thumbnail_url", "http://t2".into()),
            ("start_at", 3000i64.into()),
            ("end_at", 4000i64.into()),
        ]);
        ins.add_value(&[
            ("id", 3i64.into()),
            ("user_id", 2i64.into()),
            ("title", "t3".into()),
            ("description", "d3".into()),
            ("playlist_url", "http://p3".into()),
            ("thumbnail_url", "http://t3".into()),
            ("start_at", 5000i64.into()),
            ("end_at", 6000i64.into()),
        ]);
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let repo = LivestreamRepositoryInfra {};

    let result = repo
        .find_all_by_user_id(&mut tx, &UserId::new(1))
        .await
        .unwrap();
    assert_eq!(result.len(), 2);

    let result = repo
        .find_all_by_user_id(&mut tx, &UserId::new(2))
        .await
        .unwrap();
    assert_eq!(result.len(), 1);
    assert_eq!(*result[0].id.inner(), 3);
}
