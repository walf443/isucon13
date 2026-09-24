package mysql

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain/repository"
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
