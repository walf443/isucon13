use crate::repos::icon_repository::IconRepositoryInfra;
use crate::test_support::InsertUserSetup;
use crate::test_support::get_db_pool;
use fake::{Fake, Faker};
use isupipe_core::models::icon::CreateIcon;
use isupipe_core::models::user::{CreateUser, UserId};
use isupipe_core::repos::icon_repository::IconRepository;

#[tokio::test]
async fn success_case() {
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let user: CreateUser = Faker.fake();
    {
        InsertUserSetup { id: 1, user: &user }.insert(&mut tx).await;
    }

    let mut icon: CreateIcon = Faker.fake();
    icon.user_id = UserId::new(1);

    let repo = IconRepositoryInfra {};
    let icon_id = repo.create(&mut tx, &icon).await.unwrap();
    assert!(icon_id > 0);

    let got = repo
        .find_image_by_user_id(&mut tx, &UserId::new(1))
        .await
        .unwrap();
    assert_eq!(got, Some(icon.image));
}
