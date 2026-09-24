package repository

import (
	"context"
	"database/sql"
)

// Querier は *sqlx.DB と *sqlx.Tx の両方が満たすインターフェース。
// repository のメソッドはこれを引数で受け取ることで、トランザクションの内外どちらからでも呼べる。
type Querier interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}
