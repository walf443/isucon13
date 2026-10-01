package repository

import "context"

// TxManager は fn をトランザクション内で実行する。
// fn が error を返した場合はロールバックし、そうでなければコミットする。
type TxManager interface {
	RunInTx(ctx context.Context, fn func(q Querier) error) error
}
