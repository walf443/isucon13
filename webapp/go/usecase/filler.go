package usecase

import (
	"errors"

	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// hasPrimaryKey は ID を PrimaryKey で取り出せるエンティティ。
type hasPrimaryKey[K comparable] interface {
	PrimaryKey() K
}

// orderedBy は filled (Filler の Fill の結果) の値を、models と同じ順序 (重複も含む) で並べて返す。
func orderedBy[M hasPrimaryKey[K], K comparable, V any](models []M, filled map[K]V) []V {
	values := make([]V, len(models))
	for i, model := range models {
		values[i] = filled[model.PrimaryKey()]
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

// uniqueKeys は models それぞれから key で取り出した値を、重複を除いて最初に現れた順に返す。
// 一括で取得する repository のメソッドに渡す ID の一覧を作るのに使う。
func uniqueKeys[M any, K comparable](models []M, key func(M) K) []K {
	keys := make([]K, 0, len(models))
	seen := make(map[K]bool, len(models))
	for _, model := range models {
		k := key(model)
		if seen[k] {
			continue
		}
		seen[k] = true
		keys = append(keys, k)
	}
	return keys
}

// indexBy は models を、key で取り出した値ごとの map にする。同じ値のものが複数あれば後のもので上書きする。
func indexBy[M any, K comparable](models []M, key func(M) K) map[K]M {
	indexed := make(map[K]M, len(models))
	for _, model := range models {
		indexed[key(model)] = model
	}
	return indexed
}

// missingDetail は付随するデータが見つからなかったことを表すエラー (errMissingDetail) を返す。
// メッセージは repository.ErrNotFound のもの (移行前と同じ本文) にする。
func missingDetail() error {
	return asMissingDetail(repository.ErrNotFound)
}
