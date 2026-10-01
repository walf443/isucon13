package domain

import (
	"errors"
	"strconv"
)

// Limit は一覧取得で返す最大件数。ParseLimit で作ったものは 1 以上であることが保証される。
//
// 型としては範囲を強制できないので、domain.Limit(n) と直接変換せず ParseLimit で作ること (テストを除く)。
// 件数なので Go の慣習どおり符号付きにしている (uint64 にしても 0 は防げず、負の数からの変換が巨大な値になる)。
type Limit int64

// ErrLimitOutOfRange は Limit が許される範囲 (1 以上、上限以下) に無いことを表す。
var ErrLimitOutOfRange = errors.New("limit is out of range")

// ParseLimit は 10 進数の文字列を 1 以上 max 以下の Limit として読み取る。
// 整数でない場合は strconv のエラーを、範囲外の場合は ErrLimitOutOfRange を返す。
func ParseLimit(s string, max Limit) (Limit, error) {
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	limit := Limit(v)
	if limit < 1 || limit > max {
		return 0, ErrLimitOutOfRange
	}
	return limit, nil
}
