use crate::repos::user_repository::UserRepositoryInfra;
use crate::tables::user::UserRow;
use crate::test_support::get_db_pool;
use fake::{Fake, Faker};
use isupipe_core::models::user::{CreateUser, User};
use isupipe_core::repos::user_repository::UserRepository;

#[tokio::test]
async fn success_case() {
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let user: CreateUser = Faker.fake();

    let repo = UserRepositoryInfra {};
    let user_id = repo.create(&mut tx, &user).await.unwrap();

    let got: User = UserRow::filter(UserRow::fields().id().eq(user_id.inner()))
        .one()
        .exec(&mut tx)
        .await
        .unwrap()
        .into();

    assert_eq!(got.id, user_id);
    assert_eq!(got.name.inner(), &user.name);
    assert_eq!(got.description, Some(user.description));
    assert_eq!(got.display_name, Some(user.display_name));
}
