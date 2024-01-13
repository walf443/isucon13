use crate::repos::tag_repository::TagRepositoryInfra;
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::tag::Tag;
use isupipe_core::repos::tag_repository::TagRepository;

#[tokio::test]
async fn empty_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let tag1: Tag = Faker.fake();

    let repo = TagRepositoryInfra {};
    let result = repo.find_ids_by_name(&mut tx, &tag1.name).await.unwrap();
    assert_eq!(result.len(), 0);
}

#[tokio::test]
async fn success_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let tag1: Tag = Faker.fake();

    sqlx::query("INSERT INTO tags (id, name) VALUES (?, ?)")
        .bind(&tag1.id)
        .bind(&tag1.name)
        .execute(&mut *tx)
        .await
        .unwrap();

    let repo = TagRepositoryInfra {};
    let result = repo.find_ids_by_name(&mut tx, &tag1.name).await.unwrap();
    assert_eq!(result.len(), 1);
    let got = result.first().unwrap();
    assert_eq!(got.get(), tag1.id.get())
}
