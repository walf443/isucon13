package mysql

import (
	"context"
	"slices"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

func findTestLivestreamViewersHistories(t *testing.T, tx repository.Querier) []model.LivestreamViewersHistoryModel {
	t.Helper()
	var viewers []model.LivestreamViewersHistoryModel
	if err := tx.SelectContext(context.Background(), &viewers, "SELECT id, user_id, livestream_id, created_at FROM livestream_viewers_history ORDER BY id"); err != nil {
		t.Fatalf("failed to get livestream viewers: %v", err)
	}
	return viewers
}

func TestLivestreamViewersHistoryRepository_Create(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)
	repo := NewLivestreamViewersHistoryRepository()

	if err := repo.Create(ctx, tx, &model.LivestreamViewersHistoryModel{UserID: 1, LivestreamID: 10, CreatedAt: 100}); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	// 同じユーザ・ライブ配信でも重複して登録する (移行前と同じ)
	if err := repo.Create(ctx, tx, &model.LivestreamViewersHistoryModel{UserID: 1, LivestreamID: 10, CreatedAt: 200}); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	viewers := findTestLivestreamViewersHistories(t, tx)
	if len(viewers) != 2 {
		t.Fatalf("len(viewers) = %d, want 2", len(viewers))
	}
	for i, wantCreatedAt := range []int64{100, 200} {
		v := viewers[i]
		if v.UserID != 1 || v.LivestreamID != 10 || v.CreatedAt != wantCreatedAt {
			t.Errorf("viewers[%d] = %+v", i, v)
		}
	}
}

func TestLivestreamViewersHistoryRepository_DeleteByUserIDAndLivestreamID(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)
	repo := NewLivestreamViewersHistoryRepository()

	for _, v := range []model.LivestreamViewersHistoryModel{
		{UserID: 1, LivestreamID: 10, CreatedAt: 100},
		{UserID: 1, LivestreamID: 10, CreatedAt: 200},
		{UserID: 2, LivestreamID: 10, CreatedAt: 100},
		{UserID: 1, LivestreamID: 20, CreatedAt: 100},
	} {
		if err := repo.Create(ctx, tx, &v); err != nil {
			t.Fatalf("Create returned error: %v", err)
		}
	}

	if err := repo.DeleteByUserIDAndLivestreamID(ctx, tx, 1, 10); err != nil {
		t.Fatalf("DeleteByUserIDAndLivestreamID returned error: %v", err)
	}

	// 該当する視聴履歴は重複分も含めて全て消え、他のユーザ・ライブ配信のものは残る
	type key struct {
		userID       model.UserID
		livestreamID model.LivestreamID
	}
	var got []key
	for _, v := range findTestLivestreamViewersHistories(t, tx) {
		got = append(got, key{v.UserID, v.LivestreamID})
	}
	if want := []key{{2, 10}, {1, 20}}; !slices.Equal(got, want) {
		t.Errorf("remaining = %v, want %v", got, want)
	}
}
