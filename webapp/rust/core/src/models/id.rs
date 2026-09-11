//! kubetsu の `define_id!` で生成した汎用 ID 型。
//!
//! `Id<T, U>` は `T` (所有するモデル) で区別される `U` 型の newtype。
//! serde / fake / sqlx (MySQL) の実装は adapter crate のマクロで生成する。
kubetsu::define_id!(
    pub struct Id<T, U>;
);
kubetsu_serde::impl_serde!(Id<T, U>);
kubetsu_fake::impl_fake!(Id<T, U>);
kubetsu_sqlx::impl_sqlx!(Id<T, U>);
