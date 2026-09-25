package mysql

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

func TestLivecommentReportRepository_FindByID(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	ownerID := insertTestUser(t, tx, "alice")
	livestreamID := insertTestLivestream(t, tx, ownerID, "stream")
	livecommentID := insertTestLivecomment(t, tx, ownerID, livestreamID, "spam", 100)
	repo := NewLivecommentReportRepository("")
	want := domain.LivecommentReportModel{UserID: ownerID, LivestreamID: livestreamID, LivecommentID: livecommentID, CreatedAt: 1700000000}
	id, err := repo.Create(ctx, tx, &want)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	want.ID = id

	got, err := repo.FindByID(ctx, tx, id)
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}
	if *got != want {
		t.Errorf("report = %+v, want %+v", *got, want)
	}

	if _, err := repo.FindByID(ctx, tx, 999999); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestLivecommentReportRepository_FindAllByLivestreamID(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	ownerID := insertTestUser(t, tx, "alice")
	livestreamID := insertTestLivestream(t, tx, ownerID, "stream")
	otherID := insertTestLivestream(t, tx, ownerID, "other")
	livecommentID := insertTestLivecomment(t, tx, ownerID, livestreamID, "spam", 100)
	otherCommentID := insertTestLivecomment(t, tx, ownerID, otherID, "spam", 100)
	repo := NewLivecommentReportRepository("")
	var want []domain.LivecommentReportID
	for range 2 {
		id, err := repo.Create(ctx, tx, &domain.LivecommentReportModel{UserID: ownerID, LivestreamID: livestreamID, LivecommentID: livecommentID, CreatedAt: 100})
		if err != nil {
			t.Fatalf("Create returned error: %v", err)
		}
		want = append(want, id)
	}
	if _, err := repo.Create(ctx, tx, &domain.LivecommentReportModel{UserID: ownerID, LivestreamID: otherID, LivecommentID: otherCommentID, CreatedAt: 100}); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	got, err := repo.FindAllByLivestreamID(ctx, tx, livestreamID)
	if err != nil {
		t.Fatalf("FindAllByLivestreamID returned error: %v", err)
	}
	ids := make([]domain.LivecommentReportID, len(got))
	for i, r := range got {
		ids[i] = r.ID
	}
	// 順序は不定なので並べ替えて比べる
	slices.Sort(ids)
	if !slices.Equal(ids, want) {
		t.Errorf("IDs = %v, want %v", ids, want)
	}
}
