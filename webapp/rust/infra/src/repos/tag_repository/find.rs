use crate::repos::tag_repository::TagRepositoryInfra;
use crate::test_support::InsertTagSetup;
use crate::test_support::get_db_pool;
use fake::{Fake, Faker};
use isupipe_core::models::tag::{Tag, TagId};
use isupipe_core::repos::tag_repository::TagRepository;

#[tokio::test]
async fn not_found_case() {
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let repo = TagRepositoryInfra {};
    let tag_id: TagId = Faker.fake();
    let result = repo.find(&mut tx, &tag_id).await;
    assert!(result.is_err());
    let err_msg = format!("{}", result.unwrap_err());
    assert!(
        err_msg.contains("record not found"),
        "expected record not found, got: {}",
        err_msg
    );
}

#[tokio::test]
async fn success_case() {
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let repo = TagRepositoryInfra {};
    let tag: Tag = Faker.fake();

    {
        InsertTagSetup {
            id: *tag.id.inner(),
            tag: &tag,
        }
        .insert(&mut tx)
        .await;
    }

    let got = repo.find(&mut tx, &tag.id).await.unwrap();
    assert_eq!(got.id, tag.id);
    assert_eq!(got.name, tag.name);
}
