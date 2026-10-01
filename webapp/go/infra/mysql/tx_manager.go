package mysql

import (
	"context"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/usecase/repository"
	"gorm.io/gorm"
)

type txManager struct {
	db *gorm.DB
}

func NewTxManager(db *gorm.DB) repository.TxManager {
	return &txManager{db: db}
}

func (m *txManager) RunInTx(ctx context.Context, fn func(q repository.Querier) error) error {
	tx := m.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	// Commit 済みなら ErrTxDone が返るだけなので、エラーは無視する
	defer func() { _ = tx.Rollback().Error }()

	if err := fn(newQuerier(tx)); err != nil {
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	return nil
}
