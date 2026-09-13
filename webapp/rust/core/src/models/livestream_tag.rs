use crate::models::livestream::LivestreamId;
use crate::models::tag::TagId;

#[derive(Debug)]
pub struct LivestreamTag {
    #[allow(unused)]
    pub id: LivestreamTagId,
    pub livestream_id: LivestreamId,
    pub tag_id: TagId,
}

kubetsu::define_id!(
    #[derive(toasty::Embed)]
    pub struct LivestreamTagId(i64);
);
kubetsu_serde::impl_serde!(LivestreamTagId(i64));
kubetsu_fake::impl_fake!(LivestreamTagId(i64));
