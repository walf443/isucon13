use crate::qbey_support::bind_qbey_values;
use crate::repos::user_repository::UserRepositoryInfra;
use crate::tables::user::TABLE_USERS;
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::user::{User, UserId};
use isupipe_core::repos::user_repository::UserRepository;
use qbey::prelude::*;
use qbey_mysql::qbey;

#[tokio::test]
async fn found_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let user_id: UserId = Faker.fake();
    let repo = UserRepositoryInfra {};

    let mut user: User = Faker.fake();
    user.id = user_id.clone();
    user.display_name = Some(Faker.fake());
    let password: String = Faker.fake();
    let hashed_password = repo.hash_password(&password).unwrap();
    user.hashed_password = Some(hashed_password);
    user.description = Some(Faker.fake());

    {
        let mut ins = qbey(TABLE_USERS.table()).into_insert();
        ins.add_value(&[
            ("id", (*user.id.inner()).into()),
            ("name", user.name.inner().as_str().into()),
            ("display_name", user.display_name.as_deref().unwrap().into()),
            ("description", user.description.as_deref().unwrap().into()),
            ("password", user.hashed_password.as_deref().unwrap().into()),
        ]);
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let got = repo.find(&mut tx, &user_id).await.unwrap();
    assert!(got.is_some());
    let got = got.unwrap();
    assert_eq!(got.id, user_id);
    assert_eq!(got.name, user.name);
    assert_eq!(got.display_name, user.display_name);
    assert_eq!(got.description, user.description);
}

#[tokio::test]
async fn not_found_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let user_id: UserId = Faker.fake();
    let repo = UserRepositoryInfra {};

    let user = repo.find(&mut tx, &user_id).await.unwrap();
    assert!(user.is_none());
}
