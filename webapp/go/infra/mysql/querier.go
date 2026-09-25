package mysql

import (
	"context"
	"database/sql"

	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

// sqlxQuerier は *sqlx.DB と *sqlx.Tx の両方が満たすインターフェース。
type sqlxQuerier interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// querier は sqlxQuerier を repository.Querier として使えるようにする。
// ExecContext の戻り値を sql.Result から repository.Result にするためだけの薄いラッパー。
type querier struct {
	q sqlxQuerier
}

func newQuerier(q sqlxQuerier) repository.Querier {
	return &querier{q: q}
}

func (w *querier) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return w.q.GetContext(ctx, dest, query, args...)
}

func (w *querier) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return w.q.SelectContext(ctx, dest, query, args...)
}

func (w *querier) ExecContext(ctx context.Context, query string, args ...interface{}) (repository.Result, error) {
	return w.q.ExecContext(ctx, query, args...)
}
