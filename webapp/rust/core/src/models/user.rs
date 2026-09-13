use fake::Dummy;

#[derive(Debug, Dummy)]
pub struct User {
    pub id: UserId,
    pub name: UserName,
    pub display_name: Option<String>,
    pub description: Option<String>,
    pub hashed_password: Option<String>,
}

#[derive(Debug, Dummy, PartialEq, Clone)]
pub struct CreateUser {
    pub name: String,
    pub display_name: String,
    pub description: String,
    pub password: String,
}

kubetsu::define_id!(
    #[derive(toasty::Embed)]
    pub struct UserId(i64);
);
kubetsu_serde::impl_serde!(UserId(i64));
kubetsu_fake::impl_fake!(UserId(i64));

kubetsu::define_id!(
    #[derive(toasty::Embed)]
    pub struct UserName(String);
);
kubetsu_serde::impl_serde!(UserName(String));
kubetsu_fake::impl_fake!(UserName(String));
