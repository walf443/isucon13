use fake::Dummy;

#[derive(Debug, Dummy)]
pub struct Tag {
    pub id: TagId,
    pub name: TagName,
}

kubetsu::define_id!(
    #[derive(toasty::Embed)]
    pub struct TagId(i64);
);
kubetsu_serde::impl_serde!(TagId(i64));
kubetsu_fake::impl_fake!(TagId(i64));

kubetsu::define_id!(
    #[derive(toasty::Embed)]
    pub struct TagName(String);
);
kubetsu_serde::impl_serde!(TagName(String));
kubetsu_fake::impl_fake!(TagName(String));
