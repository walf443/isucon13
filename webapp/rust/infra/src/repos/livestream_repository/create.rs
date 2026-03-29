use crate::qbey_support::bind_qbey_values;
use crate::repos::livestream_repository::LivestreamRepositoryInfra;
use crate::tables::livestream::TABLE_LIVESTREAMS;
use crate::tables::user::TABLE_USERS;
use crate::test_support::InsertUserSetup;
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::{CreateLivestream, Livestream};
use isupipe_core::models::user::{CreateUser, UserId};
use isupipe_core::repos::livestream_repository::LivestreamRepository;
use qbey::prelude::*;
use qbey_mysql::qbey;

#[tokio::test]
async fn success_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let user: CreateUser = Faker.fake();
    {
        let mut ins = qbey(TABLE_USERS.table()).into_insert();
        ins.add_value(&InsertUserSetup { id: 1, user: &user });
        let (sql, binds) = ins.into_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let mut input: CreateLivestream = Faker.fake();
    input.user_id = UserId::new(1);

    let repo = LivestreamRepositoryInfra {};
    let livestream_id = repo.create(&mut tx, &input).await.unwrap();

    let t = &TABLE_LIVESTREAMS;
    let mut q = qbey(t.table());
    q.and_where(t.id().eq(*livestream_id.inner()));
    let (sql, binds) = q.into_sql();
    let got: Livestream = bind_qbey_values!(sqlx::query_as::<_, Livestream>(&sql), binds)
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
