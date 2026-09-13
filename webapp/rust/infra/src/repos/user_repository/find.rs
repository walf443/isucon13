use crate::repos::user_repository::UserRepositoryInfra;
use crate::tables::user::UserRow;
use crate::test_support::get_db_pool;
use fake::{Fake, Faker};
use isupipe_core::models::user::{User, UserId};
use isupipe_core::repos::user_repository::UserRepository;

#[tokio::test]
async fn found_case() {
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let user_id: UserId = Faker.fake();
    let repo = UserRepositoryInfra {};

    let mut user: User = Faker.fake();
    user.id = user_id.clone();
    user.display_name = Some(Faker.fake());
    let password: String = Faker.fake();
    let hashed_password = repo.hash_password(&password).unwrap();
    user.hashed_password = Some(hashed_password);
    user.description = Some(Faker.fake());

    UserRow::create()
        .id(&user.id)
        .name(&user.name)
        .display_name(user.display_name.as_deref().unwrap())
        .description(user.description.as_deref().unwrap())
        .hashed_password(user.hashed_password.as_deref().unwrap())
        .exec(&mut tx)
        .await
        .unwrap();

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
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let user_id: UserId = Faker.fake();
    let repo = UserRepositoryInfra {};

    let user = repo.find(&mut tx, &user_id).await.unwrap();
    assert!(user.is_none());
}
