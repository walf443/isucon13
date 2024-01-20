use crate::repos::tag_repository::TagRepositoryInfra;
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::tag::Tag;
use isupipe_core::repos::tag_repository::TagRepository;

#[tokio::test]
async fn not_found_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let repo = TagRepositoryInfra {};
    let tags = repo.find_all(&mut tx).await.unwrap();

    assert_eq!(tags.len(), 0)
}
#[tokio::test]
async fn exists_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let tag1: Tag = Faker.fake();
    let tag2: Tag = Faker.fake();

    sqlx::query("INSERT INTO tags (id, name) VALUES (?, ?), (?, ?)")
        .bind(&tag1.id)
        .bind(&tag1.name)
        .bind(&tag2.id)
        .bind(&tag2.name)
        .execute(&mut *tx)
        .await
        .unwrap();

    let repo = TagRepositoryInfra {};
    let tags = repo.find_all(&mut tx).await.unwrap();

    assert_eq!(tags.len(), 2);

    let tid = tag1.id.get();
    let got = tags.iter().find(move |t| t.id.get() == tid).unwrap();
    assert_eq!(got.name.get(), tag1.name.get());

    let tid = tag2.id.get();
    let got = tags.iter().find(move |t| t.id.get() == tid).unwrap();
    assert_eq!(got.name.get(), tag2.name.get());
}
