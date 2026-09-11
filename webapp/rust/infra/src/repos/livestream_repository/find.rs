use crate::qbey_support::bind_qbey_values;
use crate::repos::livestream_repository::LivestreamRepositoryInfra;
use crate::tables::livestream::TABLE_LIVESTREAMS;
use crate::tables::user::TABLE_USERS;
use crate::test_support::{InsertLivestreamSetup, InsertUserSetup};
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::{CreateLivestream, LivestreamId};
use isupipe_core::models::user::CreateUser;
use isupipe_core::repos::livestream_repository::LivestreamRepository;
use qbey::prelude::*;
use qbey_mysql::qbey;

#[tokio::test]
async fn found_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let user: CreateUser = Faker.fake();
    {
        let mut ins = qbey(TABLE_USERS.table()).into_insert();
        ins.add_value(&InsertUserSetup { id: 1, user: &user });
        let (sql, binds) = ins.into_sql();
        bind_qbey_values!(sqlx::query(sqlx::AssertSqlSafe(sql)), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let stream: CreateLivestream = Faker.fake();
    {
        let mut ins = qbey(TABLE_LIVESTREAMS.table()).into_insert();
        ins.add_value(&InsertLivestreamSetup {
            id: 1,
            user_id: 1,
            stream: &stream,
        });
        let (sql, binds) = ins.into_sql();
        bind_qbey_values!(sqlx::query(sqlx::AssertSqlSafe(sql)), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let repo = LivestreamRepositoryInfra {};
    let result = repo.find(&mut tx, &LivestreamId::new(1)).await.unwrap();
    assert!(result.is_some());
    let ls = result.unwrap();
    assert_eq!(*ls.id.inner(), 1);
    assert_eq!(ls.title, stream.title);
    assert_eq!(ls.start_at, stream.start_at);
    assert_eq!(ls.end_at, stream.end_at);
}

#[tokio::test]
async fn not_found_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let repo = LivestreamRepositoryInfra {};
    let result = repo.find(&mut tx, &LivestreamId::new(999)).await.unwrap();
    assert!(result.is_none());
}
