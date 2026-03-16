use crate::repos::livestream_tag_repository::LivestreamTagRepositoryInfra;
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::livestream_tag::LivestreamTag;
use isupipe_core::models::tag::TagId;
use isupipe_core::repos::livestream_tag_repository::LivestreamTagRepository;

#[tokio::test]
async fn success_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    sqlx::query(
        "INSERT INTO users (id, name, display_name, password, description) VALUES (1, 'test', 'Test', 'pw', 'desc')",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    sqlx::query(
        "INSERT INTO livestreams (id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at) VALUES (1, 1, 't1', 'd1', 'http://p', 'http://t', 1000, 2000)",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    sqlx::query("INSERT INTO tags (id, name) VALUES (1, 'tag1')")
        .execute(&mut *tx)
        .await
        .unwrap();

    let repo = LivestreamTagRepositoryInfra {};
    repo.insert(&mut tx, &LivestreamId::new(1), &TagId::new(1))
        .await
        .unwrap();

    let got: LivestreamTag =
        sqlx::query_as("SELECT * FROM livestream_tags WHERE livestream_id = ? AND tag_id = ?")
            .bind(1_i64)
            .bind(1_i64)
            .fetch_one(&mut *tx)
            .await
            .unwrap();

    assert_eq!(*got.livestream_id.inner(), 1);
    assert_eq!(*got.tag_id.inner(), 1);
}
