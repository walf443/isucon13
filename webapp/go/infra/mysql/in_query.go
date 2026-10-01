package mysql

import (
	"cmp"
	"context"
	"fmt"
	"slices"

	"github.com/isucon/isucon13/webapp/go/usecase/repository"
	"github.com/jmoiron/sqlx"
)

// maxInArgs は 1 回の IN クエリに渡す ID の最大数。
// MySQL のプレースホルダーは 1 つのクエリで 65535 個までなので、それに余裕を持たせた値にしている。
const maxInArgs = 1000

// selectIn は query (IN (?) を 1 つだけ含む) に ids を渡して実行し、結果の行を返す。
// ids は昇順に並べ、重複を除いてから渡す (同じ ids からは同じ SQL ができる)。呼び出し側の ids は変更しない。
// ids が maxInArgs を超える場合は分割して実行し、結果を連結して返す (分割すると結果の並び順は query の ORDER BY どおりにならないので、
// 並び順が必要な場合は呼び出し側で並べ直す)。
// ids が空の場合は IN () が作れないので、クエリを発行せずに空のスライスを返す。
func selectIn[T any, ID cmp.Ordered](ctx context.Context, q repository.Querier, query string, ids []ID) ([]*T, error) {
	return selectInChunked[T](ctx, q, query, ids, maxInArgs)
}

// selectInChunked は selectIn の、分割する大きさを指定できるもの (テストで小さい値を使うため)。
func selectInChunked[T any, ID cmp.Ordered](ctx context.Context, q repository.Querier, query string, ids []ID, chunkSize int) ([]*T, error) {
	sorted := slices.Compact(slices.Sorted(slices.Values(ids)))

	rows := []*T{}
	for chunk := range slices.Chunk(sorted, chunkSize) {
		expanded, params, err := sqlx.In(query, chunk)
		if err != nil {
			return nil, fmt.Errorf("failed to construct IN query: %w", err)
		}
		var chunkRows []*T
		if err := q.SelectContext(ctx, &chunkRows, expanded, params...); err != nil {
			return nil, err
		}
		rows = append(rows, chunkRows...)
	}
	return rows, nil
}
