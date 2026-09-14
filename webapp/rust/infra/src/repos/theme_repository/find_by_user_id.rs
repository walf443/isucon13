use crate::repos::theme_repository::ThemeRepositoryInfra;
use crate::test_support::InsertThemeSetup;
use crate::test_support::get_db_pool;
use fake::{Fake, Faker};
use isupipe_core::models::theme::Theme;
use isupipe_core::models::user::UserId;
use isupipe_core::repos::theme_repository::ThemeRepository;

#[tokio::test]
async fn not_found_case() {
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let repo = ThemeRepositoryInfra {};
    let user_id: UserId = Faker.fake();

    let result = repo.find_by_user_id(&mut tx, &user_id).await;
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

    let repo = ThemeRepositoryInfra {};
    let theme: Theme = Faker.fake();

    {
        InsertThemeSetup { theme: &theme }.insert(&mut tx).await;
    }

    let got = repo.find_by_user_id(&mut tx, &theme.user_id).await.unwrap();
    assert_eq!(theme.id, got.id);
    assert_eq!(theme.user_id, got.user_id);
    assert_eq!(theme.dark_mode, got.dark_mode);
}
