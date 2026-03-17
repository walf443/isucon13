use crate::qbey_support::bind_qbey_values;
use crate::repos::livestream_comment_repository::LivestreamCommentRepositoryInfra;
use crate::tables::livestream::TABLE_LIVESTREAMS;
use crate::tables::livestream_comment::TABLE_LIVECOMMENTS;
use crate::tables::user::TABLE_USERS;
use crate::test_support::{InsertLivestreamSetup, InsertUserSetup};
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::{CreateLivestream, LivestreamId};
use isupipe_core::models::livestream_comment::{CreateLivestreamComment, LivestreamComment};
use isupipe_core::models::user::{CreateUser, UserId};
use isupipe_core::repos::livestream_comment_repository::LivestreamCommentRepository;
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

    let mut input: CreateLivestreamComment = Faker.fake();
    input.user_id = UserId::new(1);
    input.livestream_id = LivestreamId::new(1);

    let repo = LivestreamCommentRepositoryInfra {};
    let comment_id = repo.create(&mut tx, &input).await.unwrap();

    let t = &TABLE_LIVECOMMENTS;
    let mut q = qbey(t.table());
    q.and_where(t.id().eq(*comment_id.inner()));
    let (sql, binds) = q.to_sql();
    let got: LivestreamComment =
        bind_qbey_values!(sqlx::query_as::<_, LivestreamComment>(&sql), binds)
            .fetch_one(&mut *tx)
            .await
            .unwrap();

    assert_eq!(got.id, comment_id);
    assert_eq!(got.user_id, input.user_id);
    assert_eq!(got.livestream_id, input.livestream_id);
    assert_eq!(got.comment, input.comment);
    assert_eq!(got.tip, input.tip);
    assert_eq!(got.created_at, input.created_at);
}
