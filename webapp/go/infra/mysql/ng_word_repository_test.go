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

func TestNGWordRepository_Matches(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)
	repo := NewNGWordRepository()

	tests := []struct {
		name    string
		comment string
		word    string
		want    bool
	}{
		{name: "contains the word", comment: "これはひどい配信だ", word: "ひどい", want: true},
		{name: "equals the word", comment: "spam", word: "spam", want: true},
		{name: "does not contain the word", comment: "楽しい配信", word: "ひどい", want: false},
		// 以下は MySQL の LIKE による判定の性質 (移行前と同じ挙動であることを確認する)
		{name: "case insensitive", comment: "THIS IS SPAM", word: "spam", want: true},
		{name: "percent in word is a wildcard", comment: "abc", word: "a%c", want: true},
		{name: "underscore in word is a wildcard", comment: "abc", word: "a_c", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.Matches(ctx, tx, tt.comment, tt.word)
			if err != nil {
				t.Fatalf("Matches returned error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Matches(%q, %q) = %v, want %v", tt.comment, tt.word, got, tt.want)
			}
		})
	}
}

func TestNGWordRepository_CreateAndFindAllByLivestreamID(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)
	repo := NewNGWordRepository()

	create := func(userID model.UserID, livestreamID model.LivestreamID, word string) model.NGWordID {
		t.Helper()
		id, err := repo.Create(ctx, tx, &model.NGWordModel{UserID: userID, LivestreamID: livestreamID, Word: word, CreatedAt: 100})
		if err != nil {
			t.Fatalf("Create returned error: %v", err)
		}
		return id
	}
	w1 := create(1, 10, "w1")
	// 登録したユーザが違っても、同じライブ配信の NG ワードは全て返す
	w2 := create(2, 10, "w2")
	create(1, 20, "other livestream")

	got, err := repo.FindAllByLivestreamID(ctx, tx, 10)
	if err != nil {
		t.Fatalf("FindAllByLivestreamID returned error: %v", err)
	}
	// ORDER BY が無いので順序には依存しない
	gotIDs := make([]model.NGWordID, len(got))
	for i, w := range got {
		gotIDs[i] = w.ID
	}
	slices.Sort(gotIDs)
	if want := []model.NGWordID{w1, w2}; !slices.Equal(gotIDs, want) {
		t.Errorf("ids = %v, want %v", gotIDs, want)
	}
	for _, w := range got {
		if w.ID == w1 {
			if want := (model.NGWordModel{ID: w1, UserID: 1, LivestreamID: 10, Word: "w1", CreatedAt: 100}); *w != want {
				t.Errorf("got %+v, want %+v", *w, want)
			}
		}
	}
}
