package mysql

import (
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
	"gorm.io/gorm"
)

// querier は GORM の DB (トランザクション) を repository.Querier として渡すための値。
// usecase は中身を見ずに repository に渡すだけで、infra の repository が gormOf で *gorm.DB を取り出す。
type querier struct {
	repository.QuerierBase
	db *gorm.DB
}

var _ repository.Querier = (*querier)(nil)

func newQuerier(db *gorm.DB) repository.Querier {
	return &querier{db: db}
}
