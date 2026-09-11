use crate::qbey_support::bind_qbey_values;
use crate::repos::user_repository::UserRepositoryInfra;
use crate::tables::user::TABLE_USERS;
use crate::test_support::InsertUserSetup;
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::user::{CreateUser, UserId};
use isupipe_core::repos::user_repository::UserRepository;
use qbey::prelude::*;
use qbey_mysql::qbey;

#[tokio::test]
async fn not_found_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let repo = UserRepositoryInfra {};
    let name: String = Faker.fake();
    let got_user = repo.find_by_name(&mut tx, &name).await.unwrap();
    assert!(got_user.is_none())
}

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
    let user_id = UserId::new(1);

    let repo = UserRepositoryInfra {};
    let got_user = repo.find_by_name(&mut tx, &user.name).await.unwrap();
    assert!(got_user.is_some());
    assert_eq!(got_user.unwrap().id, user_id)
}
