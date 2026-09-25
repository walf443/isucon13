package mysql

import (
	"context"
	"slices"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain/model"
)

// ライブ配信統計情報 (GET /api/livestream/:livestream_id/statistics) で使う集計系のクエリのテスト

func TestLivestreamRepository_FindAll(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	ownerID := insertTestUser(t, tx, "alice")
	s1 := insertTestLivestream(t, tx, ownerID, "s1")
	s2 := insertTestLivestream(t, tx, ownerID, "s2")

	livestreams, err := NewLivestreamRepository("").FindAll(ctx, tx)
	if err != nil {
		t.Fatalf("FindAll returned error: %v", err)
	}
	var ids []model.LivestreamID
	for _, l := range livestreams {
		ids = append(ids, l.ID)
		if l.ID == s1 && (l.UserID != ownerID || l.Title != "s1") {
			t.Errorf("livestream = %+v", *l)
		}
	}
	if !slices.Contains(ids, s1) || !slices.Contains(ids, s2) {
		t.Errorf("ids = %v, want to contain %d, %d", ids, s1, s2)
	}
}

func TestLivestreamStatisticsQueries(t *testing.T) {
	ctx := context.Background()
	tx := beginTestTx(t)

	ownerID := insertTestUser(t, tx, "alice")
	viewerID := insertTestUser(t, tx, "bob")
	stream := insertTestLivestream(t, tx, ownerID, "s1")
	otherStream := insertTestLivestream(t, tx, ownerID, "s2")
	emptyStream := insertTestLivestream(t, tx, ownerID, "empty")

	insertTestReaction(t, tx, viewerID, stream, "smile")
	insertTestReaction(t, tx, viewerID, stream, "heart")
	insertTestReaction(t, tx, viewerID, otherStream, "smile")

	c1 := insertTestLivecommentWithTip(t, tx, viewerID, stream, 100)
	insertTestLivecommentWithTip(t, tx, viewerID, stream, 300)
	insertTestLivecommentWithTip(t, tx, viewerID, otherStream, 1000)

	viewerRepo := NewLivestreamViewersHistoryRepository()
	for _, v := range []model.LivestreamViewersHistoryModel{
		{UserID: viewerID, LivestreamID: stream, CreatedAt: 100},
		{UserID: ownerID, LivestreamID: stream, CreatedAt: 100},
		{UserID: viewerID, LivestreamID: otherStream, CreatedAt: 100},
	} {
		if err := viewerRepo.Create(ctx, tx, &v); err != nil {
			t.Fatalf("failed to insert viewer: %v", err)
		}
	}

	reportRepo := NewLivecommentReportRepository("")
	if _, err := reportRepo.Create(ctx, tx, &model.LivecommentReportModel{UserID: ownerID, LivestreamID: stream, LivecommentID: c1, CreatedAt: 100}); err != nil {
		t.Fatalf("failed to insert report: %v", err)
	}

	reactionRepo := NewReactionRepository("")
	livecommentRepo := NewLivecommentRepository("")

	tests := []struct {
		name  string
		count func(model.LivestreamID) (int64, error)
		want  int64
		// wantEmpty はリアクション等が無いライブ配信での値
		wantEmpty int64
	}{
		{name: "reactions", count: func(id model.LivestreamID) (int64, error) { return reactionRepo.CountByLivestreamID(ctx, tx, id) }, want: 2},
		{name: "total reactions", count: func(id model.LivestreamID) (int64, error) { return reactionRepo.CountTotalByLivestreamID(ctx, tx, id) }, want: 2},
		{name: "tips", count: func(id model.LivestreamID) (int64, error) { return livecommentRepo.SumTipByLivestreamID(ctx, tx, id) }, want: 400},
		{name: "max tip", count: func(id model.LivestreamID) (int64, error) { return livecommentRepo.MaxTipByLivestreamID(ctx, tx, id) }, want: 300},
		{name: "viewers", count: func(id model.LivestreamID) (int64, error) { return viewerRepo.CountViewersByLivestreamID(ctx, tx, id) }, want: 2},
		{name: "reports", count: func(id model.LivestreamID) (int64, error) { return reportRepo.CountByLivestreamID(ctx, tx, id) }, want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.count(stream)
			if err != nil {
				t.Fatalf("returned error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
			// 集計対象が無い場合も (NULL やエラーではなく) 0
			got, err = tt.count(emptyStream)
			if err != nil || got != tt.wantEmpty {
				t.Errorf("empty livestream: got %d, %v, want %d", got, err, tt.wantEmpty)
			}
		})
	}
}
