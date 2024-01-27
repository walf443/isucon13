use crate::repos::theme_repository::ThemeRepositoryInfra;
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::theme::Theme;
use isupipe_core::repos::theme_repository::ThemeRepository;

#[tokio::test]
async fn success_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let theme: Theme = Faker.fake();

    let repo = ThemeRepositoryInfra {};
    repo.create(&mut *tx, &theme.user_id, theme.dark_mode)
        .await
        .unwrap();

    let got: Theme = sqlx::query_as("SELECT * FROM themes WHERE user_id =  ?")
        .bind(&theme.user_id)
        .fetch_one(&mut *tx)
        .await
        .unwrap();

    assert_eq!(theme.user_id, got.user_id);
    assert_eq!(theme.dark_mode, got.dark_mode);
}
