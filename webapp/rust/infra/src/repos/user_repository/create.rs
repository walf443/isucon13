use crate::repos::user_repository::UserRepositoryInfra;
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::user::{CreateUser, User};
use isupipe_core::repos::user_repository::UserRepository;
use sqlx::Acquire;

#[tokio::test]
async fn success_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let user: CreateUser = Faker.fake();

    let repo = UserRepositoryInfra {};
    let user_id = repo.create(&mut tx, &user).await.unwrap();

    let conn = tx.acquire().await.unwrap();
    let got: User = sqlx::query_as("SELECT * FROM users where id = ?")
        .bind(&user_id)
        .fetch_one(conn)
        .await
        .unwrap();
    assert_eq!(got.id, user_id);
    assert_eq!(got.name.inner(), &user.name);
    assert_eq!(got.description, Some(user.description));
    assert_eq!(got.display_name, Some(user.display_name));
}
