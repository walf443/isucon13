package mysql

import (
	"context"
	"errors"
	"testing"

	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// txManager は実際にコミットするので、他のテストと干渉しない user_id を使い、終了時に削除する。
const txManagerTestUserID = 900001

func countThemes(t *testing.T) int {
	t.Helper()
	var n int
	if err := testDB.Get(&n, "SELECT COUNT(*) FROM themes WHERE user_id = ?", txManagerTestUserID); err != nil {
		t.Fatalf("failed to count themes: %v", err)
	}
	return n
}

func TestTxManager_RunInTx(t *testing.T) {
	if testDB == nil {
		t.Skip("skipping test that requires MySQL in short mode")
	}
	ctx := context.Background()
	t.Cleanup(func() {
		_, _ = testDB.Exec("DELETE FROM themes WHERE user_id = ?", txManagerTestUserID)
	})
	insert := func(q repository.Querier) error {
		_, err := q.ExecContext(ctx, "INSERT INTO themes (user_id, dark_mode) VALUES (?, ?)", txManagerTestUserID, false)
		return err
	}

	t.Run("rolls back when fn returns error", func(t *testing.T) {
		wantErr := errors.New("boom")
		err := NewTxManager(testDB).RunInTx(ctx, func(q repository.Querier) error {
			if err := insert(q); err != nil {
				return err
			}
			return wantErr
		})
		if !errors.Is(err, wantErr) {
			t.Fatalf("err = %v, want %v", err, wantErr)
		}
		if n := countThemes(t); n != 0 {
			t.Errorf("count = %d, want 0", n)
		}
	})

	t.Run("commits when fn succeeds", func(t *testing.T) {
		if err := NewTxManager(testDB).RunInTx(ctx, insert); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if n := countThemes(t); n != 1 {
			t.Errorf("count = %d, want 1", n)
		}
	})
}
