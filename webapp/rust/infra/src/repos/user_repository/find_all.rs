use crate::qbey_support::bind_qbey_values;
use crate::repos::user_repository::UserRepositoryInfra;
use crate::tables::user::TABLE_USERS;
use crate::test_support::InsertUserSetup;
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::user::CreateUser;
use isupipe_core::repos::user_repository::UserRepository;
use qbey::prelude::*;
use qbey_mysql::qbey;

#[tokio::test]
async fn empty_case() {
    let pool = get_db_pool().await.unwrap();

    let mut tx = pool.begin().await.unwrap();
    let repo = UserRepositoryInfra {};

    let users = repo.find_all(&mut tx).await.unwrap();
    assert_eq!(users.len(), 0);
}

#[tokio::test]
async fn not_empty_case() {
    let pool = get_db_pool().await.unwrap();

    let mut tx = pool.begin().await.unwrap();
    let repo = UserRepositoryInfra {};

    let user_count = 2;
    let mut users: Vec<CreateUser> = Vec::with_capacity(user_count);
    for _ in 0..user_count {
        users.push(Faker.fake())
    }

    {
        let mut ins = qbey(TABLE_USERS.table()).into_insert();
        ins.add_value(&InsertUserSetup {
            id: 1,
            user: &users[0],
        });
        ins.add_value(&InsertUserSetup {
            id: 2,
            user: &users[1],
        });
        let (sql, binds) = ins.into_sql();
        bind_qbey_values!(sqlx::query(sqlx::AssertSqlSafe(sql)), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let users = repo.find_all(&mut tx).await.unwrap();
    assert_eq!(users.len(), user_count);
}
