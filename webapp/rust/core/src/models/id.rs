//! kubetsu の `define_id!` で生成した汎用 ID 型。
//!
//! `Id<T, U>` は `T` (所有するモデル) で区別される `U` 型の newtype。
//! serde / fake / sqlx (MySQL) の実装は adapter crate のマクロで生成する。
//! qbey の型付きカラム (`TypedCol<Id<..>>`) にそのまま渡せるよう、`qbey::Value` への変換も持つ。
kubetsu::define_id!(
    pub struct Id<T, U>;
);
kubetsu_serde::impl_serde!(Id<T, U>);
kubetsu_fake::impl_fake!(Id<T, U>);
kubetsu_sqlx::impl_sqlx!(Id<T, U>);

impl<T> From<Id<T, i64>> for qbey::Value {
    fn from(id: Id<T, i64>) -> Self {
        qbey::Value::Int(*id.inner())
    }
}

impl<T> From<Id<T, String>> for qbey::Value {
    fn from(id: Id<T, String>) -> Self {
        qbey::Value::String(id.inner().clone())
    }
}
