package mysql

import (
	"context"
	"errors"
	"maps"
	"testing"

	driver "github.com/go-sql-driver/mysql"
	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

func insertTestUser(t *testing.T, tx repository.Querier, name string) domain.UserID {
	t.Helper()
	id := insertSQL(t, tx, "INSERT INTO users (name, display_name, password, description) VALUES (?, ?, ?, ?)", name, "Display "+name, "hashed", "desc "+name)
	return domain.UserID(id)
}

func TestUserRepository_FindIDByName(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	lastID := insertSQL(t, tx, "INSERT INTO users (name, display_name, password, description) VALUES (?, ?, ?, ?)", "alice", "Alice", "x", "")
	wantID := domain.UserID(lastID)

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

	lastID := insertSQL(t, tx, "INSERT INTO users (name, display_name, password, description) VALUES (?, ?, ?, ?)", "alice", "Alice", "hashed", "hello")
	wantID := domain.UserID(lastID)

	user, err := NewUserRepository().FindByName(ctx, tx, "alice")
	if err != nil {
		t.Fatalf("FindByName returned error: %v", err)
	}
	want := domain.User{ID: wantID, Name: "alice", DisplayName: "Alice", Description: "hello", HashedPassword: "hashed"}
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

	lastID := insertSQL(t, tx, "INSERT INTO users (name, display_name, password, description) VALUES (?, ?, ?, ?)", "alice", "Alice", "hashed", "hello")
	id := domain.UserID(lastID)

	user, err := NewUserRepository().FindByID(ctx, tx, id)
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}
	want := domain.User{ID: id, Name: "alice", DisplayName: "Alice", Description: "hello", HashedPassword: "hashed"}
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

func TestUserRepository_CreateAndThemeRepository_Create(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)
	userRepo := NewUserRepository()

	id, err := userRepo.Create(ctx, tx, &domain.User{Name: "alice", DisplayName: "Alice", Description: "hello", HashedPassword: "hashed"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if err := NewThemeRepository().Create(ctx, tx, &domain.Theme{UserID: id, DarkMode: true}); err != nil {
		t.Fatalf("theme Create returned error: %v", err)
	}

	want := domain.User{ID: id, Name: "alice", DisplayName: "Alice", Description: "hello", HashedPassword: "hashed"}
	if got, err := userRepo.FindByID(ctx, tx, id); err != nil || *got != want {
		t.Errorf("FindByID = %+v, %v, want %+v", got, err, want)
	}
	if theme, err := NewThemeRepository().FindByUserID(ctx, tx, id); err != nil || theme.UserID != id || !theme.DarkMode {
		t.Errorf("theme = %+v, %v", theme, err)
	}

	// ユーザ名は UNIQUE なので重複登録はエラー (移行前と同じく 500 になる)
	if _, err := userRepo.Create(ctx, tx, &domain.User{Name: "alice", HashedPassword: "hashed"}); err == nil {
		t.Error("expected error on duplicate name")
	}
}

func TestUserRepository_FindAllByIDs(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	alice := insertTestUser(t, tx, "alice")
	bob := insertTestUser(t, tx, "bob")
	insertTestUser(t, tx, "carol")
	repo := NewUserRepository()

	// 存在しない ID は結果に含まれない
	got, err := repo.FindAllByIDs(ctx, tx, []domain.UserID{bob, alice, 999999})
	if err != nil {
		t.Fatalf("FindAllByIDs returned error: %v", err)
	}
	names := map[domain.UserID]string{}
	for _, u := range got {
		names[u.ID] = u.Name
	}
	if want := map[domain.UserID]string{alice: "alice", bob: "bob"}; !maps.Equal(names, want) {
		t.Errorf("users = %v, want %v", names, want)
	}

	// 空の場合は空のスライス
	got, err = repo.FindAllByIDs(ctx, tx, nil)
	if err != nil || got == nil || len(got) != 0 {
		t.Errorf("FindAllByIDs(nil) = %#v, %v, want empty", got, err)
	}
}

// ユーザ名が重複した場合は、GORM が翻訳せずに、ドライバのエラー (Error 1062) がそのまま返ること。
// 500 のレスポンスの本文に出るメッセージが、移行前と同じになる。
func TestUserRepository_Create_DuplicateNameReturnsDriverError(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)
	repo := NewUserRepository()

	if _, err := repo.Create(ctx, tx, &domain.User{Name: "alice", HashedPassword: "hashed"}); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	_, err := repo.Create(ctx, tx, &domain.User{Name: "alice", HashedPassword: "hashed"})

	mysqlErr, ok := errors.AsType[*driver.MySQLError](err)
	if !ok || mysqlErr.Number != 1062 {
		t.Fatalf("err = %v, want *mysql.MySQLError with number 1062", err)
	}
	if want := "Error 1062 (23000): Duplicate entry 'alice' for key 'users.uniq_user_name'"; err.Error() != want {
		t.Errorf("err = %q, want %q", err.Error(), want)
	}
}
