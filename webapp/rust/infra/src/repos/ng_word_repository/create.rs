use crate::repos::ng_word_repository::NgWordRepositoryInfra;
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::LivestreamId;
use isupipe_core::models::ng_word::{CreateNgWord, NgWord};
use isupipe_core::models::user::UserId;
use isupipe_core::repos::ng_word_repository::NgWordRepository;

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

    let input = CreateNgWord {
        user_id: UserId::new(1),
        livestream_id: LivestreamId::new(1),
        word: "badword".to_string(),
        created_at: 300,
    };

    let repo = NgWordRepositoryInfra {};
    let word_id = repo.create(&mut tx, &input).await.unwrap();

    let got: NgWord = sqlx::query_as("SELECT * FROM ng_words WHERE id = ?")
        .bind(&word_id)
        .fetch_one(&mut *tx)
        .await
        .unwrap();

    assert_eq!(got.id, word_id);
    assert_eq!(got.user_id, input.user_id);
    assert_eq!(got.livestream_id, input.livestream_id);
    assert_eq!(got.word, input.word);
    assert_eq!(got.created_at, input.created_at);
}
