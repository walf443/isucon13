use crate::qbey_support::bind_qbey_values;
use crate::repos::user_repository::UserRepositoryInfra;
use crate::tables::user::TABLE_USERS;
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::user::CreateUser;
use isupipe_core::repos::user_repository::UserRepository;
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
        ins.add_value(&[
            ("name", users[0].name.as_str().into()),
            ("display_name", users[0].display_name.as_str().into()),
            ("description", users[0].description.as_str().into()),
            ("password", users[0].password.as_str().into()),
        ]);
        ins.add_value(&[
            ("name", users[1].name.as_str().into()),
            ("display_name", users[1].display_name.as_str().into()),
            ("description", users[1].description.as_str().into()),
            ("password", users[1].password.as_str().into()),
        ]);
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let users = repo.find_all(&mut tx).await.unwrap();
    assert_eq!(users.len(), user_count);
}
