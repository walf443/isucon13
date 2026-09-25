package mysql

import (
	"context"
	"slices"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain/model"
)

func TestNGWordRepository_FindAllByUserIDAndLivestreamID(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	insert := func(userID model.UserID, livestreamID model.LivestreamID, word string, createdAt int64) model.NGWordID {
		t.Helper()
		res, err := tx.ExecContext(ctx, "INSERT INTO ng_words (user_id, livestream_id, word, created_at) VALUES (?, ?, ?, ?)", userID, livestreamID, word, createdAt)
		if err != nil {
			t.Fatalf("failed to insert ng word: %v", err)
		}
		id, _ := res.LastInsertId()
		return model.NGWordID(id)
	}
	// 作成日時の降順になることを確認するため、ID の順序とはずらす
	w1 := insert(1, 10, "w1", 300)
	w2 := insert(1, 10, "w2", 100)
	w3 := insert(1, 10, "w3", 200)
	insert(2, 10, "other user", 400)
	insert(1, 20, "other livestream", 400)
	repo := NewNGWordRepository()

	got, err := repo.FindAllByUserIDAndLivestreamID(ctx, tx, 1, 10)
	if err != nil {
		t.Fatalf("FindAllByUserIDAndLivestreamID returned error: %v", err)
	}
	gotIDs := make([]model.NGWordID, len(got))
	for i, w := range got {
		gotIDs[i] = w.ID
	}
	if want := []model.NGWordID{w1, w3, w2}; !slices.Equal(gotIDs, want) {
		t.Errorf("ids = %v, want %v", gotIDs, want)
	}
	if want := (model.NGWordModel{ID: w1, UserID: 1, LivestreamID: 10, Word: "w1", CreatedAt: 300}); *got[0] != want {
		t.Errorf("got[0] = %+v, want %+v", *got[0], want)
	}

	none, err := repo.FindAllByUserIDAndLivestreamID(ctx, tx, 3, 10)
	if err != nil {
		t.Fatalf("FindAllByUserIDAndLivestreamID returned error: %v", err)
	}
	if len(none) != 0 {
		t.Errorf("got %+v, want empty", none)
	}
}
