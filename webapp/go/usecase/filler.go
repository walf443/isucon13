package usecase

import (
	"errors"

	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// orderedBy は filled (Filler の Fill の結果) の値を、models と同じ順序 (重複も含む) で並べて返す。
// id は model から filled のキー (ID) を取り出す。
func orderedBy[M any, K comparable, V any](models []M, id func(M) K, filled map[K]V) []V {
	values := make([]V, len(models))
	for i, model := range models {
		values[i] = filled[id(model)]
	}
	return values
}

// errMissingDetail は付随するデータ (テーマ・配信者・タグなど) が見つからないことを表す。
// データ不整合なので、主となるデータの不在 (404) とは区別して 500 にする。
var errMissingDetail = errors.New("missing detail")

// missingDetailError は errMissingDetail として判定され、repository.ErrNotFound としては判定されないエラー。
// メッセージは元のエラーのまま (移行前と同じ本文) にする。
type missingDetailError struct {
	msg string
}

func (e *missingDetailError) Error() string { return e.msg }

func (e *missingDetailError) Is(target error) bool { return target == errMissingDetail }

// asMissingDetail は付随するデータを引いたときの err が repository.ErrNotFound なら、missingDetailError に置き換える。
// 呼び出し側が付随するデータの欠損を主となるデータの不在 (404) と取り違えないように、Filler はこれを通してからエラーを返す。
// ErrNotFound 以外のエラーはそのまま返す。
func asMissingDetail(err error) error {
	if errors.Is(err, repository.ErrNotFound) {
		return &missingDetailError{msg: err.Error()}
	}
	return err
}
