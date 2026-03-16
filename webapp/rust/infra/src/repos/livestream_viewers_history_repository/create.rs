use crate::qbey_support::bind_qbey_values;
use crate::repos::livestream_viewers_history_repository::LivestreamViewersHistoryRepositoryInfra;
use crate::tables::livestream::TABLE_LIVESTREAMS;
use crate::tables::user::TABLE_USERS;
use crate::test_support::{InsertLivestreamSetup, InsertUserSetup};
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::{CreateLivestream, LivestreamId};
use isupipe_core::models::livestream_viewers_history::CreateLivestreamViewersHistory;
use isupipe_core::models::user::{CreateUser, UserId};
use isupipe_core::repos::livestream_viewers_history_repository::LivestreamViewersHistoryRepository;
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

    let stream: CreateLivestream = Faker.fake();
    {
        let mut ins = qbey(TABLE_LIVESTREAMS.table()).into_insert();
        ins.add_value(&InsertLivestreamSetup { id: 1, user_id: 1, stream: &stream });
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let mut input: CreateLivestreamViewersHistory = Faker.fake();
    input.user_id = UserId::new(1);
    input.livestream_id = LivestreamId::new(1);

    let repo = LivestreamViewersHistoryRepositoryInfra {};
    repo.create(&mut tx, &input).await.unwrap();

    let count: i64 = sqlx::query_scalar(
        "SELECT COUNT(*) FROM livestream_viewers_history WHERE user_id = ? AND livestream_id = ?",
    )
    .bind(*input.user_id.inner())
    .bind(*input.livestream_id.inner())
    .fetch_one(&mut *tx)
    .await
    .unwrap();

    assert_eq!(count, 1);
}
