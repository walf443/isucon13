package mysql

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// testDB はパッケージ内のテストで共有する MySQL コンテナへの接続。
// -short 指定時は nil になり、DB を使うテストはスキップされる。
var testDB *gorm.DB

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
		testDB, err = gorm.Open(gormmysql.Open(dsn), gormConfig())
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open db: %v\n", err)
			return 1
		}
		defer func() { _ = Close(testDB) }()

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

	tx := testDB.Begin()
	if tx.Error != nil {
		t.Fatalf("failed to begin transaction: %v", tx.Error)
	}
	t.Cleanup(func() {
		_ = tx.Rollback().Error
	})
	return newQuerier(tx)
}

// recordSQL はテスト終了までに GORM が発行した SELECT / INSERT / UPDATE / DELETE (Raw / Exec も含む) の SQL (プレースホルダーのまま) を記録する。
// repository が移行前と同じカラムを読んでいる (SELECT * にしていない) ことなどを確かめるのに使う。
// 記録するのは q のトランザクションを作った *gorm.DB のコールバックなので、記録中は他のテストと並行して実行しないこと。
func recordSQL(t *testing.T, q repository.Querier) *[]string {
	t.Helper()
	db := gormOf(q)
	var recorded []string
	record := func(tx *gorm.DB) { recorded = append(recorded, tx.Statement.SQL.String()) }
	const name = "test:record_sql"
	if err := db.Callback().Query().After("gorm:query").Register(name, record); err != nil {
		t.Fatalf("failed to register query callback: %v", err)
	}
	if err := db.Callback().Create().After("gorm:create").Register(name, record); err != nil {
		t.Fatalf("failed to register create callback: %v", err)
	}
	if err := db.Callback().Delete().After("gorm:delete").Register(name, record); err != nil {
		t.Fatalf("failed to register delete callback: %v", err)
	}
	if err := db.Callback().Update().After("gorm:update").Register(name, record); err != nil {
		t.Fatalf("failed to register update callback: %v", err)
	}
	// Raw(...).Scan(...) は Row、Exec(...) は Raw のコールバックで実行される
	if err := db.Callback().Row().After("gorm:row").Register(name, record); err != nil {
		t.Fatalf("failed to register row callback: %v", err)
	}
	if err := db.Callback().Raw().After("gorm:raw").Register(name, record); err != nil {
		t.Fatalf("failed to register raw callback: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Callback().Query().Remove(name)
		_ = db.Callback().Create().Remove(name)
		_ = db.Callback().Delete().Remove(name)
		_ = db.Callback().Update().Remove(name)
		_ = db.Callback().Row().Remove(name)
		_ = db.Callback().Raw().Remove(name)
	})
	return &recorded
}
