package model

// ID は T のモデルの ID。T が異なる ID 同士は別の型になるので、取り違えるとコンパイルエラーになる。
// Rust 版で使っている kubetsu の Id<T, i64> と同じ考え方。
//
// 基底型は int64 なので、database/sql・sqlx での読み書きや JSON への変換はそのまま行える。
type ID[T any] int64
