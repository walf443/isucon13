package mysql

import (
	"cmp"
	"context"
	"slices"

	"github.com/isucon/isucon13/webapp/go/usecase/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// maxInArgs は 1 回の IN クエリに渡す ID の最大数。
// MySQL のプレースホルダーは 1 つのクエリで 65535 個までなので、それに余裕を持たせた値にしている。
const maxInArgs = 1000

// gormConfig は GORM の設定を返す。
func gormConfig() *gorm.Config {
	return &gorm.Config{
		// 移行前は SQL をログに出していなかったので、GORM のログ (遅いクエリ・エラー) も出さない
		Logger: logger.Default.LogMode(logger.Silent),
		// トランザクションは usecase が持つので、GORM が書き込みごとに開始するトランザクションは使わない
		SkipDefaultTransaction: true,
	}
}

// gormOf は repository.Querier の実体 (トランザクション) を *gorm.DB として返す。
// repository.Querier は usecase から見ると不透明な値で、infra が渡したものだけがここへ来る。
func gormOf(q repository.Querier) *gorm.DB {
	return q.(*querier).tx
}

// dbOf は gormOf と同じだが、ctx を引き継いだ *gorm.DB を返す。クエリはこれから組み立てる。
//
// 読み取りは必ず Select でカラムを列挙する (移行前の SQL と同じカラムだけを読む。テーブルにカラムが増えても結果が変わらない)。
func dbOf(ctx context.Context, q repository.Querier) *gorm.DB {
	return gormOf(q).WithContext(ctx)
}

// mapRows は行の型の一覧を domain の型の一覧にする。rows が空の場合も nil ではなく空のスライスを返す。
func mapRows[R any, D any](rows []*R, toDomain func(*R) *D) []*D {
	mapped := make([]*D, len(rows))
	for i, row := range rows {
		mapped[i] = toDomain(row)
	}
	return mapped
}

// findIn は ids を昇順・重複なしにして maxInArgs 個ずつに分け、find に渡して結果を連結して返す。
// MySQL のプレースホルダーは 1 つのクエリで 65535 個までなので、IN (...) に渡す ID が多い場合に分割する。
// 分割すると結果の並び順は find の ORDER BY どおりにならないので、並び順が必要な場合は呼び出し側で並べ直す。
// ids が空の場合は IN () が作れないので、find を呼ばずに空のスライスを返す。
func findIn[R any, ID cmp.Ordered](ids []ID, find func(chunk []ID) ([]*R, error)) ([]*R, error) {
	return findInChunked(ids, maxInArgs, find)
}

// findInChunked は findIn の、分割する大きさを指定できるもの (テストで小さい値を使うため)。
func findInChunked[R any, ID cmp.Ordered](ids []ID, chunkSize int, find func(chunk []ID) ([]*R, error)) ([]*R, error) {
	sorted := slices.Compact(slices.Sorted(slices.Values(ids)))

	rows := []*R{}
	for chunk := range slices.Chunk(sorted, chunkSize) {
		chunkRows, err := find(chunk)
		if err != nil {
			return nil, err
		}
		rows = append(rows, chunkRows...)
	}
	return rows, nil
}
