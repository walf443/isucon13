use crate::repos::icon_repository::IconRepositoryInfra;
use isupipe_core::db::get_db_pool;
use isupipe_core::models::icon::CreateIcon;
use isupipe_core::models::user::UserId;
use isupipe_core::repos::icon_repository::IconRepository;

#[tokio::test]
async fn success_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    sqlx::query(
        "INSERT INTO users (id, name, display_name, password, description) VALUES (1, 'test', 'Test', 'pw', 'desc')",
    )
    .execute(&mut *tx)
    .await
    .unwrap();

    let image_data = vec![0x89, 0x50, 0x4E, 0x47]; // PNG magic bytes
    let icon = CreateIcon {
        user_id: UserId::new(1),
        image: image_data.clone(),
    };

    let repo = IconRepositoryInfra {};
    let icon_id = repo.create(&mut tx, &icon).await.unwrap();
    assert!(icon_id > 0);

    let got = repo
        .find_image_by_user_id(&mut tx, &UserId::new(1))
        .await
        .unwrap();
    assert_eq!(got, Some(image_data));
}
