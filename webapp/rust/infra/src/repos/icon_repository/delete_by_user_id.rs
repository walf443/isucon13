use crate::qbey_support::bind_qbey_values;
use crate::qbey_support::bind_sql_values;
use crate::qbey_support::SQLValue;
use crate::repos::icon_repository::IconRepositoryInfra;
use crate::tables::icon::TABLE_ICONS;
use crate::tables::user::TABLE_USERS;
use isupipe_core::db::get_db_pool;
use isupipe_core::models::user::UserId;
use isupipe_core::repos::icon_repository::IconRepository;
use qbey_mysql::qbey;
use qbey_mysql::qbey_with;

#[tokio::test]
async fn deletes_icon() {
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
        let mut ins = qbey_with::<SQLValue>(TABLE_ICONS.table()).into_insert();
        ins.add_value(&[("id", 1i64.into()), ("user_id", 1i64.into()), ("image", vec![0x89u8, 0x50, 0x4E, 0x47].into())]);
        ins.add_value(&[("id", 2i64.into()), ("user_id", 2i64.into()), ("image", vec![0x89u8, 0x50, 0x4E, 0x47].into())]);
        let (sql, binds) = ins.to_sql();
        bind_sql_values!(sqlx::query(&sql), binds).execute(&mut *tx).await.unwrap();
    }

    let repo = IconRepositoryInfra {};
    repo.delete_by_user_id(&mut tx, &UserId::new(1))
        .await
        .unwrap();

    // user 1's icon should be gone
    let result = repo
        .find_image_by_user_id(&mut tx, &UserId::new(1))
        .await
        .unwrap();
    assert!(result.is_none());

    // user 2's icon should remain
    let result = repo
        .find_image_by_user_id(&mut tx, &UserId::new(2))
        .await
        .unwrap();
    assert!(result.is_some());
}

#[tokio::test]
async fn noop_when_no_icon() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    {
        let mut ins = qbey(TABLE_USERS.table()).into_insert();
        ins.add_value(&[("id", 1i64.into()), ("name", "test".into()), ("display_name", "Test".into()), ("password", "pw".into()), ("description", "desc".into())]);
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds).execute(&mut *tx).await.unwrap();
    }

    let repo = IconRepositoryInfra {};
    // Should not error even if no icon exists
    repo.delete_by_user_id(&mut tx, &UserId::new(1))
        .await
        .unwrap();
}
