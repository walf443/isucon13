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
	repo := NewLivecommentReportRepository()
	want := domain.LivecommentReport{UserID: ownerID, LivestreamID: livestreamID, LivecommentID: livecommentID, CreatedAt: 1700000000}
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
	repo := NewLivecommentReportRepository()
	var want []domain.LivecommentReportID
	for range 2 {
		id, err := repo.Create(ctx, tx, &domain.LivecommentReport{UserID: ownerID, LivestreamID: livestreamID, LivecommentID: livecommentID, CreatedAt: 100})
		if err != nil {
			t.Fatalf("Create returned error: %v", err)
		}
		want = append(want, id)
	}
	if _, err := repo.Create(ctx, tx, &domain.LivecommentReport{UserID: ownerID, LivestreamID: otherID, LivecommentID: otherCommentID, CreatedAt: 100}); err != nil {
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

// 報告の登録時刻は、GORM が自動で設定せずに、渡した値をそのまま保存すること (0 でも現在時刻にならない)。
func TestLivecommentReportRepository_Create_KeepsGivenCreatedAt(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	ownerID := insertTestUser(t, tx, "alice")
	livestreamID := insertTestLivestream(t, tx, ownerID, "stream")
	livecommentID := insertTestLivecomment(t, tx, ownerID, livestreamID, "spam", 100)
	repo := NewLivecommentReportRepository()

	id, err := repo.Create(ctx, tx, &domain.LivecommentReport{UserID: ownerID, LivestreamID: livestreamID, LivecommentID: livecommentID, CreatedAt: 0})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	got, err := repo.FindByID(ctx, tx, id)
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}
	if got.CreatedAt != 0 {
		t.Errorf("CreatedAt = %d, want 0", got.CreatedAt)
	}
}

// 報告数は、指定したライブ配信への報告だけを数え、ライブ配信が存在しない場合は 0 になること。
func TestLivecommentReportRepository_CountByLivestreamID(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	ownerID := insertTestUser(t, tx, "alice")
	livestreamID := insertTestLivestream(t, tx, ownerID, "stream")
	otherID := insertTestLivestream(t, tx, ownerID, "other")
	livecommentID := insertTestLivecomment(t, tx, ownerID, livestreamID, "spam", 100)
	otherCommentID := insertTestLivecomment(t, tx, ownerID, otherID, "spam", 100)
	repo := NewLivecommentReportRepository()
	for range 2 {
		if _, err := repo.Create(ctx, tx, &domain.LivecommentReport{UserID: ownerID, LivestreamID: livestreamID, LivecommentID: livecommentID, CreatedAt: 100}); err != nil {
			t.Fatalf("Create returned error: %v", err)
		}
	}
	if _, err := repo.Create(ctx, tx, &domain.LivecommentReport{UserID: ownerID, LivestreamID: otherID, LivecommentID: otherCommentID, CreatedAt: 100}); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	for _, c := range []struct {
		livestreamID domain.LivestreamID
		want         int64
	}{{livestreamID, 2}, {otherID, 1}, {999999, 0}} {
		got, err := repo.CountByLivestreamID(ctx, tx, c.livestreamID)
		if err != nil {
			t.Fatalf("CountByLivestreamID returned error: %v", err)
		}
		if got != c.want {
			t.Errorf("CountByLivestreamID(%d) = %d, want %d", c.livestreamID, got, c.want)
		}
	}
}
