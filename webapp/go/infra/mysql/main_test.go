package mysql

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
	"github.com/jmoiron/sqlx"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
)

// testDB はパッケージ内のテストで共有する MySQL コンテナへの接続。
// -short 指定時は nil になり、DB を使うテストはスキップされる。
var testDB *sqlx.DB

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		os.Exit(m.Run())
	}

	ctx := context.Background()
	container, err := mysql.Run(ctx,
		"mysql:8.0.31",
		mysql.WithDatabase("isupipe"),
		mysql.WithUsername("root"),
		mysql.WithPassword("root"),
		mysql.WithScripts(
			"../../../sql/initdb.d/00_create_database.sql",
			"../../../sql/initdb.d/10_schema.sql",
		),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start mysql container: %v\n", err)
		os.Exit(1)
	}

	code := func() int {
		defer func() {
			if err := container.Terminate(ctx); err != nil {
				fmt.Fprintf(os.Stderr, "failed to terminate mysql container: %v\n", err)
			}
		}()

		dsn, err := container.ConnectionString(ctx, "parseTime=true")
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to get connection string: %v\n", err)
			return 1
		}
		testDB, err = sqlx.Open("mysql", dsn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open db: %v\n", err)
			return 1
		}
		defer testDB.Close()

		return m.Run()
	}()
	os.Exit(code)
}

// beginTestTx はテスト終了時にロールバックされるトランザクションを、repository.Querier として返す。
// repository は Querier を受け取るので、これを渡せばテスト間でデータが干渉しない。
func beginTestTx(t *testing.T) repository.Querier {
	t.Helper()
	if testDB == nil {
		t.Skip("skipping test that requires MySQL in short mode")
	}

	tx, err := testDB.BeginTxx(context.Background(), nil)
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}
	t.Cleanup(func() {
		_ = tx.Rollback()
	})
	return newQuerier(tx)
}
