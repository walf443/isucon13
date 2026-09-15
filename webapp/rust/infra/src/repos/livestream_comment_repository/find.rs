use crate::qbey_support::bind_qbey_values;
use crate::repos::livestream_comment_repository::LivestreamCommentRepositoryInfra;
use crate::tables::livestream::TABLE_LIVESTREAMS;
use crate::tables::livestream_comment::TABLE_LIVECOMMENTS;
use crate::tables::user::TABLE_USERS;
use crate::test_support::{InsertCommentSetup, InsertLivestreamSetup, InsertUserSetup};
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::livestream::CreateLivestream;
use isupipe_core::models::livestream_comment::{CreateLivestreamComment, LivestreamCommentId};
use isupipe_core::models::user::CreateUser;
use isupipe_core::repos::livestream_comment_repository::LivestreamCommentRepository;
use qbey::prelude::*;
use qbey_mysql::qbey;

#[tokio::test]
async fn found_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let user: CreateUser = Faker.fake();
    {
        let mut ins = qbey(TABLE_USERS.table()).into_insert();
        ins.add_value(&InsertUserSetup { id: 1, user: &user });
        let (sql, binds) = ins.into_sql();
        bind_qbey_values!(sqlx::query(sqlx::AssertSqlSafe(sql)), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let stream: CreateLivestream = Faker.fake();
    {
        let mut ins = qbey(TABLE_LIVESTREAMS.table()).into_insert();
        ins.add_value(&InsertLivestreamSetup {
            id: 1,
            user_id: 1,
            stream: &stream,
        });
        let (sql, binds) = ins.into_sql();
        bind_qbey_values!(sqlx::query(sqlx::AssertSqlSafe(sql)), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    {
        let comment: CreateLivestreamComment = CreateLivestreamComment {
            comment: "hello".to_string(),
            tip: 100,
            created_at: 500,
            ..Faker.fake()
        };
        let mut ins = qbey(TABLE_LIVECOMMENTS.table()).into_insert();
        ins.add_value(&InsertCommentSetup {
            id: 1,
            user_id: 1,
            livestream_id: 1,
            comment: &comment,
        });
        let (sql, binds) = ins.into_sql();
        bind_qbey_values!(sqlx::query(sqlx::AssertSqlSafe(sql)), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let repo = LivestreamCommentRepositoryInfra {};
    let result = repo
        .find(&mut tx, &LivestreamCommentId::new(1))
        .await
        .unwrap();
    assert!(result.is_some());
    let comment = result.unwrap();
    assert_eq!(*comment.id.inner(), 1);
    assert_eq!(comment.comment, "hello");
    assert_eq!(comment.tip, 100);
}

#[tokio::test]
async fn not_found_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let repo = LivestreamCommentRepositoryInfra {};
    let result = repo
        .find(&mut tx, &LivestreamCommentId::new(999))
        .await
        .unwrap();
    assert!(result.is_none());
}
