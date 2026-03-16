use crate::qbey_support::bind_qbey_values;
use crate::repos::icon_repository::IconRepositoryInfra;
use crate::tables::user::TABLE_USERS;
use crate::test_support::InsertUserSetup;
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::icon::CreateIcon;
use isupipe_core::models::user::{CreateUser, UserId};
use isupipe_core::repos::icon_repository::IconRepository;
use qbey_mysql::qbey;

#[tokio::test]
async fn success_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let user: CreateUser = Faker.fake();
    {
        let mut ins = qbey(TABLE_USERS.table()).into_insert();
        ins.add_value(&InsertUserSetup { id: 1, user: &user });
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let mut icon: CreateIcon = Faker.fake();
    icon.user_id = UserId::new(1);

    let repo = IconRepositoryInfra {};
    let icon_id = repo.create(&mut tx, &icon).await.unwrap();
    assert!(icon_id > 0);

    let got = repo
        .find_image_by_user_id(&mut tx, &UserId::new(1))
        .await
        .unwrap();
    assert_eq!(got, Some(icon.image));
}
