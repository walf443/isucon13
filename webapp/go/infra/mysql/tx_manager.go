package mysql

import (
	"context"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain/repository"
	"github.com/jmoiron/sqlx"
)

type txManager struct {
	db *sqlx.DB
}

func NewTxManager(db *sqlx.DB) repository.TxManager {
	return &txManager{db: db}
}

func (m *txManager) RunInTx(ctx context.Context, fn func(q repository.Querier) error) error {
	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	return nil
}
