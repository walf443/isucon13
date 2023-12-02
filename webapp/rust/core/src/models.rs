pub mod tag;

#[derive(Debug, serde::Serialize)]
struct Tag {
    id: i64,
    name: String,
}
