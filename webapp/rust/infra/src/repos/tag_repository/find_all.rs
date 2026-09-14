use crate::repos::tag_repository::TagRepositoryInfra;
use crate::test_support::InsertTagSetup;
use crate::test_support::get_db_pool;
use fake::{Fake, Faker};
use isupipe_core::models::tag::Tag;
use isupipe_core::repos::tag_repository::TagRepository;

#[tokio::test]
async fn not_found_case() {
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let repo = TagRepositoryInfra {};
    let tags = repo.find_all(&mut tx).await.unwrap();

    assert_eq!(tags.len(), 0)
}
#[tokio::test]
async fn exists_case() {
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let tag1: Tag = Faker.fake();
    let tag2: Tag = Faker.fake();

    {
        InsertTagSetup {
            id: *tag1.id.inner(),
            tag: &tag1,
        }
        .insert(&mut tx)
        .await;
        InsertTagSetup {
            id: *tag2.id.inner(),
            tag: &tag2,
        }
        .insert(&mut tx)
        .await;
    }

    let repo = TagRepositoryInfra {};
    let tags = repo.find_all(&mut tx).await.unwrap();

    assert_eq!(tags.len(), 2);

    let tid = tag1.id.clone();
    let got = tags.iter().find(move |t| t.id == tid).unwrap();
    assert_eq!(got.name, tag1.name);

    let tid = tag2.id.clone();
    let got = tags.iter().find(move |t| t.id == tid).unwrap();
    assert_eq!(got.name, tag2.name);
}
