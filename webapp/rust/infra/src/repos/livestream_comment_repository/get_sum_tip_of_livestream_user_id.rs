use crate::repos::livestream_comment_repository::LivestreamCommentRepositoryInfra;
use crate::test_support::get_db_pool;
use crate::test_support::{InsertCommentSetup, InsertLivestreamSetup, InsertUserSetup};
use fake::{Fake, Faker};
use isupipe_core::models::livestream::CreateLivestream;
use isupipe_core::models::livestream_comment::CreateLivestreamComment;
use isupipe_core::models::user::{CreateUser, UserId};
use isupipe_core::repos::livestream_comment_repository::LivestreamCommentRepository;

#[tokio::test]
async fn empty_case() {
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

    let repo = LivestreamCommentRepositoryInfra {};
    let total = repo
        .get_sum_tip_of_livestream_user_id(&mut tx, &UserId::new(1))
        .await
        .unwrap();
    assert_eq!(total, 0);
}

#[tokio::test]
async fn not_empty_case() {
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let alice: CreateUser = Faker.fake();
    let bob: CreateUser = Faker.fake();
    {
        InsertUserSetup {
            id: 1,
            user: &alice,
        }
        .insert(&mut tx)
        .await;
        InsertUserSetup { id: 2, user: &bob }.insert(&mut tx).await;
    }

    let stream1: CreateLivestream = Faker.fake();
    let stream2: CreateLivestream = Faker.fake();
    let stream3: CreateLivestream = Faker.fake();
    {
        InsertLivestreamSetup {
            id: 1,
            user_id: 1,
            stream: &stream1,
        }
        .insert(&mut tx)
        .await;
        InsertLivestreamSetup {
            id: 2,
            user_id: 1,
            stream: &stream2,
        }
        .insert(&mut tx)
        .await;
        InsertLivestreamSetup {
            id: 3,
            user_id: 2,
            stream: &stream3,
        }
        .insert(&mut tx)
        .await;
    }

    {
        let c1: CreateLivestreamComment = CreateLivestreamComment {
            tip: 100,
            created_at: 1,
            comment: "a".to_string(),
            ..Faker.fake()
        };
        let c2: CreateLivestreamComment = CreateLivestreamComment {
            tip: 200,
            created_at: 2,
            comment: "b".to_string(),
            ..Faker.fake()
        };
        let c3: CreateLivestreamComment = CreateLivestreamComment {
            tip: 50,
            created_at: 3,
            comment: "c".to_string(),
            ..Faker.fake()
        };
        let c4: CreateLivestreamComment = CreateLivestreamComment {
            tip: 300,
            created_at: 4,
            comment: "d".to_string(),
            ..Faker.fake()
        };
        InsertCommentSetup {
            id: 1,
            user_id: 2,
            livestream_id: 1,
            comment: &c1,
        }
        .insert(&mut tx)
        .await;
        InsertCommentSetup {
            id: 2,
            user_id: 2,
            livestream_id: 1,
            comment: &c2,
        }
        .insert(&mut tx)
        .await;
        InsertCommentSetup {
            id: 3,
            user_id: 2,
            livestream_id: 2,
            comment: &c3,
        }
        .insert(&mut tx)
        .await;
        InsertCommentSetup {
            id: 4,
            user_id: 1,
            livestream_id: 3,
            comment: &c4,
        }
        .insert(&mut tx)
        .await;
    }

    let repo = LivestreamCommentRepositoryInfra {};

    // alice owns livestream 1 and 2, tips: 100+200+50=350
    let total = repo
        .get_sum_tip_of_livestream_user_id(&mut tx, &UserId::new(1))
        .await
        .unwrap();
    assert_eq!(total, 350);

    // bob owns livestream 3, tips: 300
    let total = repo
        .get_sum_tip_of_livestream_user_id(&mut tx, &UserId::new(2))
        .await
        .unwrap();
    assert_eq!(total, 300);
}
