use crate::repos::tag_repository::TagRepositoryInfra;
use crate::test_support::InsertTagSetup;
use crate::test_support::get_db_pool;
use fake::{Fake, Faker};
use isupipe_core::models::tag::Tag;
use isupipe_core::repos::tag_repository::TagRepository;

#[tokio::test]
async fn empty_case() {
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let tag1: Tag = Faker.fake();

    let repo = TagRepositoryInfra {};
    let result = repo.find_ids_by_name(&mut tx, &tag1.name).await.unwrap();
    assert_eq!(result.len(), 0);
}

#[tokio::test]
async fn success_case() {
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let tag1: Tag = Faker.fake();

    {
        InsertTagSetup {
            id: *tag1.id.inner(),
            tag: &tag1,
        }
        .insert(&mut tx)
        .await;
    }

    let repo = TagRepositoryInfra {};
    let result = repo.find_ids_by_name(&mut tx, &tag1.name).await.unwrap();
    assert_eq!(result.len(), 1);
    let got = result.first().unwrap();
    assert_eq!(got, &tag1.id)
}
