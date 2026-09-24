package mysql

import (
	"context"
	"errors"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain/model"
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

func TestUserRepository_FindByName(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	res, err := tx.ExecContext(ctx, "INSERT INTO users (name, display_name, password, description) VALUES (?, ?, ?, ?)", "alice", "Alice", "hashed", "hello")
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}
	wantID, _ := res.LastInsertId()

	user, err := NewUserRepository().FindByName(ctx, tx, "alice")
	if err != nil {
		t.Fatalf("FindByName returned error: %v", err)
	}
	want := model.UserModel{ID: wantID, Name: "alice", DisplayName: "Alice", Description: "hello", HashedPassword: "hashed"}
	if *user != want {
		t.Errorf("user = %+v, want %+v", *user, want)
	}
}

func TestUserRepository_FindByName_NotFound(t *testing.T) {
	tx := beginTestTx(t)

	_, err := NewUserRepository().FindByName(context.Background(), tx, "nobody")
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestUserRepository_FindByID(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	res, err := tx.ExecContext(ctx, "INSERT INTO users (name, display_name, password, description) VALUES (?, ?, ?, ?)", "alice", "Alice", "hashed", "hello")
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}
	id, _ := res.LastInsertId()

	user, err := NewUserRepository().FindByID(ctx, tx, id)
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}
	want := model.UserModel{ID: id, Name: "alice", DisplayName: "Alice", Description: "hello", HashedPassword: "hashed"}
	if *user != want {
		t.Errorf("user = %+v, want %+v", *user, want)
	}
}

func TestUserRepository_FindByID_NotFound(t *testing.T) {
	tx := beginTestTx(t)

	_, err := NewUserRepository().FindByID(context.Background(), tx, 999999)
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}
