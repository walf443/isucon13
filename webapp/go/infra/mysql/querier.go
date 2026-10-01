package mysql

import (
	"context"
	"database/sql"
	"strings"

	"github.com/isucon/isucon13/webapp/go/usecase/repository"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	"gorm.io/gorm"
)

// querier は GORM のトランザクションを repository.Querier として渡すための値。
//
// GORM に移行している途中なので、まだ移行していない repository のために、同じトランザクション (*sql.Tx) の上で動く
// sqlx のラッパーも持つ。移行した repository は dbOf で tx を使い、移行前の repository は GetContext などを使う。
// 全ての repository を移行したら、このラッパーと repository.Querier のメソッドは無くす。
type querier struct {
	tx     *gorm.DB
	legacy *sqlx.Tx
}

func newQuerier(tx *gorm.DB) repository.Querier {
	sqlTx, ok := tx.Statement.ConnPool.(*sql.Tx)
	if !ok {
		panic("newQuerier: tx is not a transaction started by GORM")
	}
	return &querier{
		tx: tx,
		// sqlx.DB が使うものと同じ、db タグでカラムに対応づける設定
		legacy: &sqlx.Tx{Tx: sqlTx, Mapper: reflectx.NewMapperFunc("db", strings.ToLower)},
	}
}

func (w *querier) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return w.legacy.GetContext(ctx, dest, query, args...)
}

func (w *querier) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return w.legacy.SelectContext(ctx, dest, query, args...)
}

func (w *querier) ExecContext(ctx context.Context, query string, args ...interface{}) (repository.Result, error) {
	return w.legacy.ExecContext(ctx, query, args...)
}
