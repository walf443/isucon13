package mysql

import (
	"context"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

func insertTestUser(t *testing.T, tx repository.Querier, name string) domain.UserID {
	t.Helper()
	res, err := tx.ExecContext(context.Background(), "INSERT INTO users (name, display_name, password, description) VALUES (?, ?, ?, ?)", name, "Display "+name, "hashed", "desc "+name)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}
	id, _ := res.LastInsertId()
	return domain.UserID(id)
}

func insertTestTheme(t *testing.T, tx repository.Querier, userID domain.UserID, darkMode bool) domain.ThemeID {
	t.Helper()
	res, err := tx.ExecContext(context.Background(), "INSERT INTO themes (user_id, dark_mode) VALUES (?, ?)", userID, darkMode)
	if err != nil {
		t.Fatalf("failed to insert theme: %v", err)
	}
	id, _ := res.LastInsertId()
	return domain.ThemeID(id)
}

func TestFillUsers(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)
	fallback := []byte("fallback")

	aliceID := insertTestUser(t, tx, "alice")
	aliceThemeID := insertTestTheme(t, tx, aliceID, true)
	if _, err := tx.ExecContext(ctx, "INSERT INTO icons (user_id, image) VALUES (?, ?)", aliceID, []byte("alice icon")); err != nil {
		t.Fatalf("failed to insert icon: %v", err)
	}
	bobID := insertTestUser(t, tx, "bob")
	bobThemeID := insertTestTheme(t, tx, bobID, false)

	userModels := []*domain.UserModel{
		// 入力の順序が保たれることを確認するため、ID の降順で渡す
		{ID: bobID, Name: "bob", DisplayName: "Display bob", Description: "desc bob"},
		{ID: aliceID, Name: "alice", DisplayName: "Display alice", Description: "desc alice"},
	}

	users, err := fillUsers(ctx, tx, userModels, domain.HashIcon(fallback))
	if err != nil {
		t.Fatalf("fillUsers returned error: %v", err)
	}

	want := []domain.User{
		{
			ID:          bobID,
			Name:        "bob",
			DisplayName: "Display bob",
			Description: "desc bob",
			Theme:       domain.ThemeModel{ID: bobThemeID, UserID: bobID, DarkMode: false},
			IconHash:    domain.HashIcon(fallback),
		},
		{
			ID:          aliceID,
			Name:        "alice",
			DisplayName: "Display alice",
			Description: "desc alice",
			Theme:       domain.ThemeModel{ID: aliceThemeID, UserID: aliceID, DarkMode: true},
			IconHash:    domain.HashIcon([]byte("alice icon")),
		},
	}
	if len(users) != len(want) {
		t.Fatalf("len(users) = %d, want %d", len(users), len(want))
	}
	for i := range want {
		if *users[i] != want[i] {
			t.Errorf("users[%d] = %+v, want %+v", i, *users[i], want[i])
		}
	}
}

func TestFillUsers_Empty(t *testing.T) {
	tx := beginTestTx(t)

	users, err := fillUsers(context.Background(), tx, nil, "")
	if err != nil {
		t.Fatalf("fillUsers returned error: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("len(users) = %d, want 0", len(users))
	}
}
