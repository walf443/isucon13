use crate::repos::livestream_comment_repository::LivestreamCommentRepositoryInfra;
use crate::test_support::get_db_pool;
use crate::test_support::{InsertCommentSetup, InsertLivestreamSetup, InsertUserSetup};
use fake::{Fake, Faker};
use isupipe_core::models::livestream::CreateLivestream;
use isupipe_core::models::livestream_comment::{CreateLivestreamComment, LivestreamCommentId};
use isupipe_core::models::user::CreateUser;
use isupipe_core::repos::livestream_comment_repository::LivestreamCommentRepository;

#[tokio::test]
async fn found_case() {
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let user: CreateUser = Faker.fake();
    {
        InsertUserSetup { id: 1, user: &user }.insert(&mut tx).await;
    }

    let stream: CreateLivestream = Faker.fake();
    {
        InsertLivestreamSetup {
            id: 1,
            user_id: 1,
            stream: &stream,
        }
        .insert(&mut tx)
        .await;
    }

    {
        let comment: CreateLivestreamComment = CreateLivestreamComment {
            comment: "hello".to_string(),
            tip: 100,
            created_at: 500,
            ..Faker.fake()
        };
        InsertCommentSetup {
            id: 1,
            user_id: 1,
            livestream_id: 1,
            comment: &comment,
        }
        .insert(&mut tx)
        .await;
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
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let repo = LivestreamCommentRepositoryInfra {};
    let result = repo
        .find(&mut tx, &LivestreamCommentId::new(999))
        .await
        .unwrap();
    assert!(result.is_none());
}
