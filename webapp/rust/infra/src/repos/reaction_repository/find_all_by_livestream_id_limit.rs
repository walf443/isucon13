use crate::qbey_support::bind_qbey_values;
use crate::repos::reaction_repository::ReactionRepositoryInfra;
use crate::tables::livestream::TABLE_LIVESTREAMS;
use crate::tables::reaction::TABLE_REACTIONS;
use crate::tables::user::TABLE_USERS;
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::repos::reaction_repository::ReactionRepository;
use qbey_mysql::qbey;

#[tokio::test]
async fn returns_limited_rows() {
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
        let mut ins = qbey(TABLE_REACTIONS.table()).into_insert();
        ins.add_value(&[("id", 1i64.into()), ("user_id", 1i64.into()), ("livestream_id", 1i64.into()), ("emoji_name", "\u{1f44d}".into()), ("created_at", 100i64.into())]);
        ins.add_value(&[("id", 2i64.into()), ("user_id", 1i64.into()), ("livestream_id", 1i64.into()), ("emoji_name", "\u{2764}\u{fe0f}".into()), ("created_at", 300i64.into())]);
        ins.add_value(&[("id", 3i64.into()), ("user_id", 1i64.into()), ("livestream_id", 1i64.into()), ("emoji_name", "\u{1f600}".into()), ("created_at", 200i64.into())]);
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds).execute(&mut *tx).await.unwrap();
    }

    let repo = ReactionRepositoryInfra {};
    let result = repo
        .find_all_by_livestream_id_limit(&mut tx, &LivestreamId::new(1), 2)
        .await
        .unwrap();
    assert_eq!(result.len(), 2);
    assert_eq!(result[0].created_at, 300);
    assert_eq!(result[1].created_at, 200);
}
