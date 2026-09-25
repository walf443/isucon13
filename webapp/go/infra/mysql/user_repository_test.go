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
	lastID, _ := res.LastInsertId()
	wantID := model.UserID(lastID)

	id, err := NewUserRepository("").FindIDByName(ctx, tx, "alice")
	if err != nil {
		t.Fatalf("FindIDByName returned error: %v", err)
	}
	if id != wantID {
		t.Errorf("id = %d, want %d", id, wantID)
	}
}

func TestUserRepository_FindIDByName_NotFound(t *testing.T) {
	tx := beginTestTx(t)

	_, err := NewUserRepository("").FindIDByName(context.Background(), tx, "nobody")
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
	lastID, _ := res.LastInsertId()
	wantID := model.UserID(lastID)

	user, err := NewUserRepository("").FindByName(ctx, tx, "alice")
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

	_, err := NewUserRepository("").FindByName(context.Background(), tx, "nobody")
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
	lastID, _ := res.LastInsertId()
	id := model.UserID(lastID)

	user, err := NewUserRepository("").FindByID(ctx, tx, id)
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

	_, err := NewUserRepository("").FindByID(context.Background(), tx, 999999)
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestUserRepository_CreateAndThemeRepository_Create(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)
	userRepo := NewUserRepository(model.IconHash([]byte("fallback")))

	id, err := userRepo.Create(ctx, tx, &model.UserModel{Name: "alice", DisplayName: "Alice", Description: "hello", HashedPassword: "hashed"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if err := NewThemeRepository().Create(ctx, tx, &model.ThemeModel{UserID: id, DarkMode: true}); err != nil {
		t.Fatalf("theme Create returned error: %v", err)
	}

	got, err := userRepo.FindWithDetailsByID(ctx, tx, id)
	if err != nil {
		t.Fatalf("FindWithDetailsByID returned error: %v", err)
	}
	if got.ID != id || got.Name != "alice" || got.DisplayName != "Alice" || got.Description != "hello" || got.Theme.UserID != id || !got.Theme.DarkMode {
		t.Errorf("user = %+v", got)
	}
	if userModel, err := userRepo.FindByID(ctx, tx, id); err != nil || userModel.HashedPassword != "hashed" {
		t.Errorf("FindByID = %+v, %v", userModel, err)
	}

	// ユーザ名は UNIQUE なので重複登録はエラー (移行前と同じく 500 になる)
	if _, err := userRepo.Create(ctx, tx, &model.UserModel{Name: "alice", HashedPassword: "hashed"}); err == nil {
		t.Error("expected error on duplicate name")
	}
}
