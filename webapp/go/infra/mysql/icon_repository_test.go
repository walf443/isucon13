package mysql

import (
	"bytes"
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

func TestIconRepository_FindImageByUserID(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	want := []byte{0x89, 'P', 'N', 'G', 0x00, 0xff}
	execSQL(t, tx, "INSERT INTO icons (user_id, image) VALUES (?, ?)", 42, want)

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

func TestIconRepository_FindAllByUserIDs(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	alice := insertTestUser(t, tx, "alice")
	// bob はアイコンが未登録
	bob := insertTestUser(t, tx, "bob")
	repo := NewIconRepository()
	if _, err := repo.Create(ctx, tx, alice, []byte("alice icon")); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	got, err := repo.FindAllByUserIDs(ctx, tx, []domain.UserID{alice, bob})
	if err != nil {
		t.Fatalf("FindAllByUserIDs returned error: %v", err)
	}
	if len(got) != 1 || got[0].UserID != alice || string(got[0].Image) != "alice icon" || got[0].ID == 0 {
		t.Errorf("icons = %+v, want only alice's icon", got)
	}

	got, err = repo.FindAllByUserIDs(ctx, tx, nil)
	if err != nil || got == nil || len(got) != 0 {
		t.Errorf("FindAllByUserIDs(nil) = %#v, %v, want empty", got, err)
	}
}

// 同じユーザのアイコンが複数ある場合 (データ不整合) は、全て ID の昇順で返す。
func TestIconRepository_FindAllByUserIDs_OrderedByID(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	alice := insertTestUser(t, tx, "alice")
	bob := insertTestUser(t, tx, "bob")
	// ユーザの ID の順ではなく、アイコンの ID の順に並ぶこと
	var ids []domain.IconID
	for _, c := range []struct {
		userID domain.UserID
		image  string
	}{{bob, "bob"}, {alice, "alice old"}, {alice, "alice new"}} {
		id := insertSQL(t, tx, "INSERT INTO icons (user_id, image) VALUES (?, ?)", c.userID, []byte(c.image))
		ids = append(ids, domain.IconID(id))
	}

	got, err := NewIconRepository().FindAllByUserIDs(ctx, tx, []domain.UserID{alice, bob})
	if err != nil {
		t.Fatalf("FindAllByUserIDs returned error: %v", err)
	}
	gotIDs := make([]domain.IconID, len(got))
	for i, icon := range got {
		gotIDs[i] = icon.ID
	}
	if !slices.Equal(gotIDs, ids) {
		t.Errorf("icon IDs = %v, want %v", gotIDs, ids)
	}
}

// 同じユーザのアイコンが複数ある場合 (データ不整合) に、FindImageByUserID (GetIcon の画像) と FindAllByUserIDs (アイコンのハッシュ) が
// 同じ行を指すこと。
func TestIconRepository_FindImageByUserIDAndFindAllByUserIDs_AgreeOnDuplicatedIcons(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	alice := insertTestUser(t, tx, "alice")
	for _, image := range []string{"first", "second"} {
		execSQL(t, tx, "INSERT INTO icons (user_id, image) VALUES (?, ?)", alice, []byte(image))
	}
	repo := NewIconRepository()

	image, err := repo.FindImageByUserID(ctx, tx, alice)
	if err != nil {
		t.Fatalf("FindImageByUserID returned error: %v", err)
	}
	icons, err := repo.FindAllByUserIDs(ctx, tx, []domain.UserID{alice})
	if err != nil || len(icons) == 0 {
		t.Fatalf("FindAllByUserIDs = %v, %v", icons, err)
	}
	// 先頭 (ID が最小) のものが、どちらでも使われる行
	if string(image) != "first" || string(icons[0].Image) != "first" {
		t.Errorf("FindImageByUserID = %q, first of FindAllByUserIDs = %q, want both %q", image, icons[0].Image, "first")
	}
}
