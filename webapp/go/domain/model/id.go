package model

import "strconv"

// ID は T のモデルの ID。T が異なる ID 同士は別の型になるので、取り違えるとコンパイルエラーになる。
// Rust 版で使っている kubetsu の Id<T, i64> と同じ考え方。
//
// 基底型は int64 なので、database/sql・sqlx での読み書きや JSON への変換はそのまま行える。
// (encoding.TextUnmarshaler を実装すると JSON の数値から読み込めなくなるので、文字列からの変換は ParseID で行う)
type ID[T any] int64

// ParseID は 10 進数の文字列を T の ID として読み取る。
func ParseID[T any](s string) (ID[T], error) {
	// 移行前の handler と同じく strconv.Atoi で解釈する
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return ID[T](v), nil
}
