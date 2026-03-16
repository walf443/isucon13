use crate::qbey_support::bind_qbey_values;
use crate::repos::livestream_repository::LivestreamRepositoryInfra;
use crate::tables::livestream::TABLE_LIVESTREAMS;
use crate::tables::user::TABLE_USERS;
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::repos::livestream_repository::LivestreamRepository;
use qbey_mysql::qbey;

#[tokio::test]
async fn found_case() {
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

    {
        let mut ins = qbey(TABLE_LIVESTREAMS.table()).into_insert();
        ins.add_value(&[
            ("id", 1i64.into()),
            ("user_id", 1i64.into()),
            ("title", "title1".into()),
            ("description", "desc1".into()),
            ("playlist_url", "http://p".into()),
            ("thumbnail_url", "http://t".into()),
            ("start_at", 1000i64.into()),
            ("end_at", 2000i64.into()),
        ]);
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let repo = LivestreamRepositoryInfra {};
    let result = repo.find(&mut tx, &LivestreamId::new(1)).await.unwrap();
    assert!(result.is_some());
    let ls = result.unwrap();
    assert_eq!(*ls.id.inner(), 1);
    assert_eq!(ls.title, "title1");
    assert_eq!(ls.start_at, 1000);
    assert_eq!(ls.end_at, 2000);
}

#[tokio::test]
async fn not_found_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let repo = LivestreamRepositoryInfra {};
    let result = repo.find(&mut tx, &LivestreamId::new(999)).await.unwrap();
    assert!(result.is_none());
}
