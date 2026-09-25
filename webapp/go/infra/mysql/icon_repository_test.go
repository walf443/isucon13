package mysql

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

func TestIconRepository_FindImageByUserID(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	want := []byte{0x89, 'P', 'N', 'G', 0x00, 0xff}
	if _, err := tx.ExecContext(ctx, "INSERT INTO icons (user_id, image) VALUES (?, ?)", 42, want); err != nil {
		t.Fatalf("failed to insert icon: %v", err)
	}

	image, err := NewIconRepository().FindImageByUserID(ctx, tx, 42)
	if err != nil {
		t.Fatalf("FindImageByUserID returned error: %v", err)
	}
	if !bytes.Equal(image, want) {
		t.Errorf("image = %v, want %v", image, want)
	}
}

func TestIconRepository_FindImageByUserID_NotFound(t *testing.T) {
	tx := beginTestTx(t)

	_, err := NewIconRepository().FindImageByUserID(context.Background(), tx, 999999)
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestIconRepository_Create(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)
	repo := NewIconRepository()

	want := []byte("new icon")
	id, err := repo.Create(ctx, tx, 42, want)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if id == 0 {
		t.Error("id should not be zero")
	}

	image, err := repo.FindImageByUserID(ctx, tx, 42)
	if err != nil {
		t.Fatalf("FindImageByUserID returned error: %v", err)
	}
	if !bytes.Equal(image, want) {
		t.Errorf("image = %q, want %q", image, want)
	}
}

func TestIconRepository_DeleteByUserID(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)
	repo := NewIconRepository()

	for _, userID := range []domain.UserID{42, 43} {
		if _, err := repo.Create(ctx, tx, userID, []byte("icon")); err != nil {
			t.Fatalf("Create returned error: %v", err)
		}
	}

	if err := repo.DeleteByUserID(ctx, tx, 42); err != nil {
		t.Fatalf("DeleteByUserID returned error: %v", err)
	}

	if _, err := repo.FindImageByUserID(ctx, tx, 42); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound for deleted user", err)
	}
	// 他のユーザのアイコンは消えない
	if _, err := repo.FindImageByUserID(ctx, tx, 43); err != nil {
		t.Errorf("icon of other user should remain: %v", err)
	}
}
