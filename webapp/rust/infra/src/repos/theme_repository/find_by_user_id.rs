use crate::qbey_support::bind_qbey_values;
use crate::repos::theme_repository::ThemeRepositoryInfra;
use crate::tables::theme::TABLE_THEMES;
use crate::test_support::InsertThemeSetup;
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::theme::Theme;
use isupipe_core::models::user::UserId;
use isupipe_core::repos::theme_repository::ThemeRepository;
use qbey::prelude::*;
use qbey_mysql::qbey;

#[tokio::test]
async fn not_found_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let repo = ThemeRepositoryInfra {};
    let user_id: UserId = Faker.fake();

    let result = repo.find_by_user_id(&mut tx, &user_id).await;
    assert!(result.is_err());
    let err_msg = format!("{}", result.unwrap_err());
    assert!(
        err_msg.contains("no rows returned"),
        "expected RowNotFound, got: {}",
        err_msg
    );
}

#[tokio::test]
async fn success_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let repo = ThemeRepositoryInfra {};
    let theme: Theme = Faker.fake();

    {
        let mut ins = qbey(TABLE_THEMES.table()).into_insert();
        ins.add_value(&InsertThemeSetup { theme: &theme });
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let got = repo.find_by_user_id(&mut tx, &theme.user_id).await.unwrap();
    assert_eq!(theme.id, got.id);
    assert_eq!(theme.user_id, got.user_id);
    assert_eq!(theme.dark_mode, got.dark_mode);
}
