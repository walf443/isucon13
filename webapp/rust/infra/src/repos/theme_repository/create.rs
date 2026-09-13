use crate::repos::theme_repository::ThemeRepositoryInfra;
use crate::tables::theme::ThemeRow;
use crate::test_support::get_db_pool;
use fake::{Fake, Faker};
use isupipe_core::models::theme::Theme;
use isupipe_core::repos::theme_repository::ThemeRepository;

#[tokio::test]
async fn success_case() {
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let theme: Theme = Faker.fake();

    let repo = ThemeRepositoryInfra {};
    repo.create(&mut tx, &theme.user_id, theme.dark_mode)
        .await
        .unwrap();

    let got: Theme = ThemeRow::all()
        .filter(ThemeRow::fields().user_id().eq(*theme.user_id.inner()))
        .one()
        .exec(&mut tx)
        .await
        .unwrap()
        .into();

    assert_eq!(theme.user_id, got.user_id);
    assert_eq!(theme.dark_mode, got.dark_mode);
}
