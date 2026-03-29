use crate::qbey_support::bind_qbey_values;
use crate::repos::theme_repository::ThemeRepositoryInfra;
use crate::tables::theme::TABLE_THEMES;
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::theme::Theme;
use isupipe_core::repos::theme_repository::ThemeRepository;
use qbey::prelude::*;
use qbey_mysql::qbey;

#[tokio::test]
async fn success_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let theme: Theme = Faker.fake();

    let repo = ThemeRepositoryInfra {};
    repo.create(&mut tx, &theme.user_id, theme.dark_mode)
        .await
        .unwrap();

    let t = &TABLE_THEMES;
    let mut q = qbey(t.table());
    q.and_where(t.user_id().eq(*theme.user_id.inner()));
    let (sql, binds) = q.into_sql();
    let got: Theme = bind_qbey_values!(sqlx::query_as::<_, Theme>(&sql), binds)
        .fetch_one(&mut *tx)
        .await
        .unwrap();

    assert_eq!(theme.user_id, got.user_id);
    assert_eq!(theme.dark_mode, got.dark_mode);
}
