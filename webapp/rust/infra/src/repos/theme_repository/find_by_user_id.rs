use crate::repos::theme_repository::ThemeRepositoryInfra;
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::theme::Theme;
use isupipe_core::models::user::UserId;
use isupipe_core::repos::theme_repository::ThemeRepository;

#[tokio::test]
#[should_panic(expected = "RowNotFound")]
async fn not_found_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let repo = ThemeRepositoryInfra {};
    let user_id: UserId = Faker.fake();

    repo.find_by_user_id(&mut tx, &user_id).await.unwrap();
}

#[tokio::test]
async fn success_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let repo = ThemeRepositoryInfra {};
    let theme: Theme = Faker.fake();

    sqlx::query("INSERT INTO themes (id, user_id, dark_mode) VALUES (?, ?, ?)")
        .bind(&theme.id)
        .bind(&theme.user_id)
        .bind(theme.dark_mode)
        .execute(&mut *tx)
        .await
        .unwrap();

    let got = repo.find_by_user_id(&mut tx, &theme.user_id).await.unwrap();
    assert_eq!(theme.id.get(), got.id.get());
    assert_eq!(theme.user_id.get(), got.user_id.get());
    assert_eq!(theme.dark_mode, got.dark_mode);
}
