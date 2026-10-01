package mysql

import (
	"testing"

	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// このファイルは、テストのデータを SQL で直接登録・確認するための補助。
// repository を通さずに登録することで、テスト対象の repository の実装に依存しないデータを用意できる。

// insertSQL は INSERT 文を tx で実行し、登録した行の ID (LAST_INSERT_ID) を返す。
func insertSQL(t *testing.T, tx repository.Querier, query string, args ...any) int64 {
	t.Helper()
	db := gormOf(tx)
	if err := db.Exec(query, args...).Error; err != nil {
		t.Fatalf("failed to execute %q: %v", query, err)
	}
	// 同じトランザクション (同じ接続) の中なので、直前の INSERT の ID が返る
	var id int64
	if err := db.Raw("SELECT LAST_INSERT_ID()").Scan(&id).Error; err != nil {
		t.Fatalf("failed to get the last insert id: %v", err)
	}
	return id
}

// execSQL は SQL を tx で実行する。登録した行の ID が要らない場合に使う。
func execSQL(t *testing.T, tx repository.Querier, query string, args ...any) {
	t.Helper()
	if err := gormOf(tx).Exec(query, args...).Error; err != nil {
		t.Fatalf("failed to execute %q: %v", query, err)
	}
}

// scanSQL は SELECT 文を tx で実行し、結果を dest (スカラー・スカラーのスライス・行の型のスライス) に読み込む。
func scanSQL(t *testing.T, tx repository.Querier, dest any, query string, args ...any) {
	t.Helper()
	if err := gormOf(tx).Raw(query, args...).Scan(dest).Error; err != nil {
		t.Fatalf("failed to query %q: %v", query, err)
	}
}
