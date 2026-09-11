use crate::qbey_support::bind_qbey_values;
use crate::repos::user_repository::UserRepositoryInfra;
use crate::tables::user::TABLE_USERS;
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::user::{CreateUser, User};
use isupipe_core::repos::user_repository::UserRepository;
use qbey::prelude::*;
use qbey_mysql::qbey;
use sqlx::Acquire;

#[tokio::test]
async fn success_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let user: CreateUser = Faker.fake();

    let repo = UserRepositoryInfra {};
    let user_id = repo.create(&mut tx, &user).await.unwrap();

    let t = &TABLE_USERS;
    let mut q = qbey(t.table());
    q.and_where(t.id().eq(*user_id.inner()));
    q.select(&t.default_cols());
    let (sql, binds) = q.into_sql();
    let conn = tx.acquire().await.unwrap();
    let got: User = bind_qbey_values!(sqlx::query_as::<_, User>(sqlx::AssertSqlSafe(sql)), binds)
        .fetch_one(conn)
        .await
        .unwrap();

    assert_eq!(got.id, user_id);
    assert_eq!(got.name.inner(), &user.name);
    assert_eq!(got.description, Some(user.description));
    assert_eq!(got.display_name, Some(user.display_name));
}
