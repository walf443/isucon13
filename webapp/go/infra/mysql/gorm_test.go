package mysql

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// find に渡る ID は重複を除いて昇順に並び、その順に分割されること。
func TestFindInChunked_SendsSortedDistinctIDsInChunks(t *testing.T) {
	var chunks [][]int
	rows, err := findInChunked([]int{3, 1, 3, 5, 2, 1, 4}, 2, func(chunk []int) ([]*int, error) {
		chunks = append(chunks, slices.Clone(chunk))
		found := make([]*int, len(chunk))
		for i := range chunk {
			found[i] = &chunk[i]
		}
		return found, nil
	})
	if err != nil {
		t.Fatalf("findInChunked returned error: %v", err)
	}

	if want := [][]int{{1, 2}, {3, 4}, {5}}; !reflect.DeepEqual(chunks, want) {
		t.Errorf("chunks = %v, want %v", chunks, want)
	}
	// 分割した結果は、渡した順に連結される
	got := make([]int, len(rows))
	for i, r := range rows {
		got[i] = *r
	}
	if want := []int{1, 2, 3, 4, 5}; !slices.Equal(got, want) {
		t.Errorf("rows = %v, want %v", got, want)
	}
}

func TestFindInChunked_EmptyDoesNotCallFind(t *testing.T) {
	rows, err := findInChunked([]int(nil), 2, func([]int) ([]*int, error) {
		t.Error("find should not be called")
		return nil, nil
	})
	if err != nil || rows == nil || len(rows) != 0 {
		t.Errorf("findInChunked(nil) = %#v, %v, want empty non-nil slice", rows, err)
	}
}

func TestFindInChunked_ReturnsFindError(t *testing.T) {
	boom := errors.New("boom")
	calls := 0
	_, err := findInChunked([]int{1, 2, 3}, 1, func([]int) ([]*int, error) {
		calls++
		return nil, boom
	})
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want %v", err, boom)
	}
	// 最初のエラーで打ち切る
	if calls != 1 {
		t.Errorf("find calls = %d, want 1", calls)
	}
}

// MySQL のプレースホルダーの上限 (65535) を超える数の ID でも、GORM の IN ? を分割して引けること。
func TestTagRepository_FindAllByIDs_MoreIDsThanMySQLPlaceholderLimit(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	ownerID := insertTestUser(t, tx, "alice")
	livestreamID := insertTestLivestream(t, tx, ownerID, "stream")
	a := insertTestTag(t, tx, livestreamID, "A")
	b := insertTestTag(t, tx, livestreamID, "B")
	ids := []domain.TagID{a.ID}
	// 存在しない ID で水増しする
	for id := domain.TagID(1_000_000); len(ids) < 70_000; id++ {
		ids = append(ids, id)
	}
	ids = append(ids, b.ID)

	got, err := NewTagRepository().FindAllByIDs(ctx, tx, ids)
	if err != nil {
		t.Fatalf("FindAllByIDs returned error: %v", err)
	}
	names := map[string]bool{}
	for _, tag := range got {
		names[tag.Name] = true
	}
	if len(got) != 2 || !names["A"] || !names["B"] {
		t.Errorf("tags = %+v, want A and B", got)
	}
}

// GORM に移行した repository が、移行前と同じカラムを読む SQL を発行していること (SELECT * にしない)。
func TestMigratedRepositories_IssueExplicitColumnSQL(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)
	sqls := recordSQL(t, tx)

	userID := insertTestUser(t, tx, "alice")
	tags := NewTagRepository()
	themes := NewThemeRepository()
	icons := NewIconRepository()
	// 記録するのは、ここから後に発行された SQL
	*sqls = nil

	calls := []struct {
		name string
		call func() error
		want string
	}{
		{"tag FindAll", func() error { _, err := tags.FindAll(ctx, tx); return err },
			"SELECT id, name FROM `tags`"},
		{"tag FindIDsByName", func() error { _, err := tags.FindIDsByName(ctx, tx, "x"); return err },
			"SELECT `id` FROM `tags` WHERE name = ?"},
		{"tag FindAllByIDs", func() error { _, err := tags.FindAllByIDs(ctx, tx, []domain.TagID{1, 2}); return err },
			"SELECT id, name FROM `tags` WHERE id IN (?,?)"},
		{"theme FindByUserID", func() error { _, err := themes.FindByUserID(ctx, tx, userID); return err },
			"SELECT id, user_id, dark_mode FROM `themes` WHERE user_id = ? LIMIT ?"},
		{"theme FindAllByUserIDs", func() error { _, err := themes.FindAllByUserIDs(ctx, tx, []domain.UserID{userID}); return err },
			"SELECT id, user_id, dark_mode FROM `themes` WHERE user_id IN (?)"},
		{"theme Create", func() error { return themes.Create(ctx, tx, &domain.Theme{UserID: userID, DarkMode: true}) },
			"INSERT INTO `themes` (`user_id`,`dark_mode`) VALUES (?,?)"},
		{"icon FindImageByUserID", func() error { _, err := icons.FindImageByUserID(ctx, tx, userID); return err },
			"SELECT `image` FROM `icons` WHERE user_id = ? ORDER BY `icons`.`id` LIMIT ?"},
		{"icon FindAllByUserIDs", func() error {
			_, err := icons.FindAllByUserIDs(ctx, tx, []domain.UserID{userID, userID + 1})
			return err
		},
			"SELECT id, user_id, image FROM `icons` WHERE user_id IN (?,?) ORDER BY id"},
		{"icon Create", func() error { _, err := icons.Create(ctx, tx, userID, []byte("img")); return err },
			"INSERT INTO `icons` (`user_id`,`image`) VALUES (?,?)"},
		{"icon DeleteByUserID", func() error { return icons.DeleteByUserID(ctx, tx, userID) },
			"DELETE FROM `icons` WHERE user_id = ?"},
	}
	for _, c := range calls {
		*sqls = nil
		err := c.call()
		// theme / icon の FindXxx は、まだ無いので ErrNotFound になるが、SQL は発行されている
		if err != nil && !errors.Is(err, repository.ErrNotFound) {
			t.Fatalf("%s returned error: %v", c.name, err)
		}
		if len(*sqls) != 1 || (*sqls)[0] != c.want {
			t.Errorf("%s: SQL = %q, want [%q]", c.name, *sqls, c.want)
		}
	}
}
