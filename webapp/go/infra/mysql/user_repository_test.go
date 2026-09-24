package mysql

import (
	"context"
	"errors"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

func TestUserRepository_FindIDByName(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	res, err := tx.ExecContext(ctx, "INSERT INTO users (name, display_name, password, description) VALUES (?, ?, ?, ?)", "alice", "Alice", "x", "")
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}
	wantID, _ := res.LastInsertId()

	id, err := NewUserRepository().FindIDByName(ctx, tx, "alice")
	if err != nil {
		t.Fatalf("FindIDByName returned error: %v", err)
	}
	if id != wantID {
		t.Errorf("id = %d, want %d", id, wantID)
	}
}

func TestUserRepository_FindIDByName_NotFound(t *testing.T) {
	tx := beginTestTx(t)

	_, err := NewUserRepository().FindIDByName(context.Background(), tx, "nobody")
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}
