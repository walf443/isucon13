package usecase

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

type fakeTxManager struct {
	// runs は RunInTx が呼ばれた回数
	runs int
}

func (m *fakeTxManager) RunInTx(ctx context.Context, fn func(q repository.Querier) error) error {
	m.runs++
	return fn(nil)
}
