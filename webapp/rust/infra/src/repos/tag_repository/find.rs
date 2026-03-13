use crate::repos::tag_repository::TagRepositoryInfra;
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::tag::{Tag, TagId};
use isupipe_core::repos::tag_repository::TagRepository;

#[tokio::test]
async fn not_found_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let repo = TagRepositoryInfra {};
    let tag_id: TagId = Faker.fake();
    let result = repo.find(&mut tx, &tag_id).await;
    assert!(result.is_err());
    let err_msg = format!("{}", result.unwrap_err());
    assert!(err_msg.contains("no rows returned"), "expected RowNotFound, got: {}", err_msg);
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
    assert_eq!(got.id, tag.id);
    assert_eq!(got.name, tag.name);
}
