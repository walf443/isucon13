package mysql

import (
	"context"
	"errors"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

func insertTestUser(t *testing.T, tx repository.Querier, name string) model.UserID {
	t.Helper()
	res, err := tx.ExecContext(context.Background(), "INSERT INTO users (name, display_name, password, description) VALUES (?, ?, ?, ?)", name, "Display "+name, "hashed", "desc "+name)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}
	id, _ := res.LastInsertId()
	return model.UserID(id)
}

func insertTestTheme(t *testing.T, tx repository.Querier, userID model.UserID, darkMode bool) model.ThemeID {
	t.Helper()
	res, err := tx.ExecContext(context.Background(), "INSERT INTO themes (user_id, dark_mode) VALUES (?, ?)", userID, darkMode)
	if err != nil {
		t.Fatalf("failed to insert theme: %v", err)
	}
	id, _ := res.LastInsertId()
	return model.ThemeID(id)
}

func TestUserRepository_FindWithDetails(t *testing.T) {
	fallback := []byte("fallback")

	tests := []struct {
		name         string
		icon         []byte
		wantIconHash model.IconHash
	}{
		{
			name:         "with registered icon",
			icon:         []byte("icon"),
			wantIconHash: model.HashIcon([]byte("icon")),
		},
		{
			name:         "without icon uses fallback",
			icon:         nil,
			wantIconHash: model.HashIcon(fallback),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			tx := beginTestTx(t)

			userID := insertTestUser(t, tx, "alice")
			themeID := insertTestTheme(t, tx, userID, true)
			if tt.icon != nil {
				if _, err := tx.ExecContext(ctx, "INSERT INTO icons (user_id, image) VALUES (?, ?)", userID, tt.icon); err != nil {
					t.Fatalf("failed to insert icon: %v", err)
				}
			}

			want := model.User{
				ID:          userID,
				Name:        "alice",
				DisplayName: "Display alice",
				Description: "desc alice",
				Theme:       model.ThemeModel{ID: themeID, UserID: userID, DarkMode: true},
				IconHash:    tt.wantIconHash,
			}
			repo := NewUserRepository(model.HashIcon(fallback))

			byName, err := repo.FindWithDetailsByName(ctx, tx, "alice")
			if err != nil {
				t.Fatalf("FindWithDetailsByName returned error: %v", err)
			}
			if *byName != want {
				t.Errorf("FindWithDetailsByName = %+v, want %+v", *byName, want)
			}

			byID, err := repo.FindWithDetailsByID(ctx, tx, userID)
			if err != nil {
				t.Fatalf("FindWithDetailsByID returned error: %v", err)
			}
			if *byID != want {
				t.Errorf("FindWithDetailsByID = %+v, want %+v", *byID, want)
			}
		})
	}
}

func TestUserRepository_FindWithDetails_UserNotFound(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)
	repo := NewUserRepository("")

	if _, err := repo.FindWithDetailsByName(ctx, tx, "nobody"); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("FindWithDetailsByName err = %v, want ErrNotFound", err)
	}
	if _, err := repo.FindWithDetailsByID(ctx, tx, 999999); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("FindWithDetailsByID err = %v, want ErrNotFound", err)
	}
}

func TestUserRepository_FindWithDetails_ThemeNotFound(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)
	insertTestUser(t, tx, "alice")

	// テーマ欠損はデータ不整合なので、ユーザ不在 (ErrNotFound) とは区別する
	_, err := NewUserRepository("").FindWithDetailsByName(ctx, tx, "alice")
	if err == nil {
		t.Fatal("expected error")
	}
	if errors.Is(err, repository.ErrNotFound) {
		t.Errorf("err = %v, should not be ErrNotFound", err)
	}
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

	userModels := []*model.UserModel{
		// 入力の順序が保たれることを確認するため、ID の降順で渡す
		{ID: bobID, Name: "bob", DisplayName: "Display bob", Description: "desc bob"},
		{ID: aliceID, Name: "alice", DisplayName: "Display alice", Description: "desc alice"},
	}

	users, err := fillUsers(ctx, tx, userModels, model.HashIcon(fallback))
	if err != nil {
		t.Fatalf("fillUsers returned error: %v", err)
	}

	want := []model.User{
		{
			ID:          bobID,
			Name:        "bob",
			DisplayName: "Display bob",
			Description: "desc bob",
			Theme:       model.ThemeModel{ID: bobThemeID, UserID: bobID, DarkMode: false},
			IconHash:    model.HashIcon(fallback),
		},
		{
			ID:          aliceID,
			Name:        "alice",
			DisplayName: "Display alice",
			Description: "desc alice",
			Theme:       model.ThemeModel{ID: aliceThemeID, UserID: aliceID, DarkMode: true},
			IconHash:    model.HashIcon([]byte("alice icon")),
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
