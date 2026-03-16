use crate::repos::livestream_repository::LivestreamRepositoryInfra;
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::{CreateLivestream, Livestream};
use isupipe_core::models::user::UserId;
use isupipe_core::repos::livestream_repository::LivestreamRepository;

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

    let input = CreateLivestream {
        user_id: UserId::new(1),
        title: "my stream".to_string(),
        description: "a description".to_string(),
        playlist_url: "http://playlist".to_string(),
        thumbnail_url: "http://thumb".to_string(),
        start_at: 1000,
        end_at: 2000,
    };

    let repo = LivestreamRepositoryInfra {};
    let livestream_id = repo.create(&mut tx, &input).await.unwrap();

    let got: Livestream = sqlx::query_as("SELECT * FROM livestreams WHERE id = ?")
        .bind(&livestream_id)
        .fetch_one(&mut *tx)
        .await
        .unwrap();

    assert_eq!(got.id, livestream_id);
    assert_eq!(got.user_id, input.user_id);
    assert_eq!(got.title, input.title);
    assert_eq!(got.description, input.description);
    assert_eq!(got.playlist_url, input.playlist_url);
    assert_eq!(got.thumbnail_url, input.thumbnail_url);
    assert_eq!(got.start_at, input.start_at);
    assert_eq!(got.end_at, input.end_at);
}
