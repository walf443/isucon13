use crate::repos::tag_repository::TagRepositoryInfra;
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::tag::{Tag, TagId};
use isupipe_core::repos::tag_repository::TagRepository;

#[tokio::test]
#[should_panic(expected = "RowNotFound")]
async fn not_found_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let repo = TagRepositoryInfra {};
    let tag_id: TagId = Faker.fake();
    repo.find(&mut tx, &tag_id).await.unwrap();
}

#[tokio::test]
async fn success_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let repo = TagRepositoryInfra {};
    let tag: Tag = Faker.fake();

    sqlx::query("INSERT INTO tags (id, name) VALUES (?, ?)")
        .bind(&tag.id)
        .bind(&tag.name)
        .execute(&mut *tx)
        .await
        .unwrap();

    let got = repo.find(&mut tx, &tag.id).await.unwrap();
    assert_eq!(got.id.get(), tag.id.get());
    assert_eq!(got.name, tag.name);
}
