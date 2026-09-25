package repository

import (
	"context"
)

// Querier は repository がクエリを実行する先 (DB またはトランザクション)。
// repository のメソッドはこれを引数で受け取ることで、トランザクションの内外どちらからでも呼べる。
// トランザクションの境界をコード上で明示し、repository を 1 メソッドずつテストしやすくするため、あえて引数で渡している。
type Querier interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...interface{}) (Result, error)
}

// Result は更新系のクエリの結果。database/sql の sql.Result のうち、repository が使うものだけを持つ。
type Result interface {
	LastInsertId() (int64, error)
	RowsAffected() (int64, error)
}
