package usecase

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

func newTestUserStatisticsRepos() (*fakeUserRepository, *fakeLivestreamRepository, *fakeLivecommentRepository, *fakeReactionRepository, *fakeLivestreamViewersHistoryRepository) {
	userRepo := &fakeUserRepository{
		findByName: func(_ context.Context, _ repository.Querier, name string) (*model.UserModel, error) {
			if name != "bob" {
				return nil, fmt.Errorf("unexpected user name %q", name)
			}
			return &model.UserModel{ID: 2, Name: "bob"}, nil
		},
		findAll: func(context.Context, repository.Querier) ([]*model.UserModel, error) {
			return []*model.UserModel{
				{ID: 1, Name: "alice"},
				{ID: 2, Name: "bob"},
				{ID: 3, Name: "carol"},
			}, nil
		},
	}
	livestreamRepo := &fakeLivestreamRepository{
		findAllByUserID: func(_ context.Context, _ repository.Querier, userID model.UserID) ([]*model.LivestreamModel, error) {
			if userID != 2 {
				return nil, fmt.Errorf("unexpected livestream owner %d", userID)
			}
			return []*model.LivestreamModel{{ID: 10, UserID: 2}, {ID: 11, UserID: 2}}, nil
		},
	}
	// スコアは リアクション数 + チップ合計: alice 30, bob 30, carol 5
	tips := map[model.UserID]int64{1: 20, 2: 25, 3: 5}
	livecomments := map[model.LivestreamID][]*model.LivecommentModel{
		10: {{Tip: 10}, {Tip: 0}},
		11: {{Tip: 15}},
	}
	livecommentRepo := &fakeLivecommentRepository{
		sumTipByLivestreamOwnerID: func(_ context.Context, _ repository.Querier, userID model.UserID) (int64, error) {
			return tips[userID], nil
		},
		findAllByLivestreamID: func(_ context.Context, _ repository.Querier, livestreamID model.LivestreamID) ([]*model.LivecommentModel, error) {
			return livecomments[livestreamID], nil
		},
	}
	reactionCounts := map[model.UserID]int64{1: 10, 2: 5, 3: 0}
	// 配信者名で集計するクエリは、統計対象のユーザ (bob) で呼ばれる
	checkOwnerName := func(name string) error {
		if name != "bob" {
			return fmt.Errorf("unexpected owner name %q", name)
		}
		return nil
	}
	reactionRepo := &fakeReactionRepository{
		countByLivestreamOwnerID: func(_ context.Context, _ repository.Querier, userID model.UserID) (int64, error) {
			return reactionCounts[userID], nil
		},
		countByLivestreamOwnerName: func(_ context.Context, _ repository.Querier, name string) (int64, error) {
			return 5, checkOwnerName(name)
		},
		findFavoriteEmojiByLivestreamOwnerName: func(_ context.Context, _ repository.Querier, name string) (string, error) {
			return "smile", checkOwnerName(name)
		},
	}
	viewerCounts := map[model.LivestreamID]int64{10: 3, 11: 4}
	viewerRepo := &fakeLivestreamViewersHistoryRepository{
		countByLivestreamID: func(_ context.Context, _ repository.Querier, livestreamID model.LivestreamID) (int64, error) {
			return viewerCounts[livestreamID], nil
		},
	}
	return userRepo, livestreamRepo, livecommentRepo, reactionRepo, viewerRepo
}

func TestStatisticsUsecase_FindUserStatistics(t *testing.T) {
	userRepo, livestreamRepo, livecommentRepo, reactionRepo, viewerRepo := newTestUserStatisticsRepos()
	u := NewStatisticsUsecase(&fakeTxManager{}, userRepo, livestreamRepo, livecommentRepo, reactionRepo, viewerRepo, &fakeLivecommentReportRepository{})

	got, err := u.FindUserStatistics(context.Background(), "bob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := model.UserStatistics{
		// alice と同点だが、ユーザ名の昇順で後ろの bob が上位になる
		Rank:              1,
		ViewersCount:      7,
		TotalReactions:    5,
		TotalLivecomments: 3,
		TotalTip:          25,
		FavoriteEmoji:     "smile",
	}
	if *got != want {
		t.Errorf("got %+v, want %+v", *got, want)
	}
}

func TestStatisticsUsecase_FindUserStatistics_NoReactions(t *testing.T) {
	userRepo, livestreamRepo, livecommentRepo, reactionRepo, viewerRepo := newTestUserStatisticsRepos()
	reactionRepo.findFavoriteEmojiByLivestreamOwnerName = func(context.Context, repository.Querier, string) (string, error) {
		return "", repository.ErrNotFound
	}
	u := NewStatisticsUsecase(&fakeTxManager{}, userRepo, livestreamRepo, livecommentRepo, reactionRepo, viewerRepo, &fakeLivecommentReportRepository{})

	// リアクションが無い場合はエラーにせず、お気に入り絵文字を空にする (移行前と同じ)
	got, err := u.FindUserStatistics(context.Background(), "bob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.FavoriteEmoji != "" {
		t.Errorf("favorite emoji = %q, want empty", got.FavoriteEmoji)
	}
}

func TestStatisticsUsecase_FindUserStatistics_Errors(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name    string
		modify  func(*fakeUserRepository, *fakeLivestreamRepository, *fakeLivecommentRepository, *fakeReactionRepository, *fakeLivestreamViewersHistoryRepository)
		wantErr error
		wantMsg string
	}{
		{
			name: "user not found",
			modify: func(ur *fakeUserRepository, _ *fakeLivestreamRepository, _ *fakeLivecommentRepository, _ *fakeReactionRepository, _ *fakeLivestreamViewersHistoryRepository) {
				ur.findByName = func(context.Context, repository.Querier, string) (*model.UserModel, error) {
					return nil, repository.ErrNotFound
				}
			},
			wantErr: ErrUserNotFound,
		},
		{
			name: "get user fails",
			modify: func(ur *fakeUserRepository, _ *fakeLivestreamRepository, _ *fakeLivecommentRepository, _ *fakeReactionRepository, _ *fakeLivestreamViewersHistoryRepository) {
				ur.findByName = func(context.Context, repository.Querier, string) (*model.UserModel, error) { return nil, boom }
			},
			wantErr: boom,
			wantMsg: "failed to get user: boom",
		},
		{
			name: "get users fails",
			modify: func(ur *fakeUserRepository, _ *fakeLivestreamRepository, _ *fakeLivecommentRepository, _ *fakeReactionRepository, _ *fakeLivestreamViewersHistoryRepository) {
				ur.findAll = func(context.Context, repository.Querier) ([]*model.UserModel, error) { return nil, boom }
			},
			wantErr: boom,
			wantMsg: "failed to get users: boom",
		},
		{
			name: "count reactions fails",
			modify: func(_ *fakeUserRepository, _ *fakeLivestreamRepository, _ *fakeLivecommentRepository, rr *fakeReactionRepository, _ *fakeLivestreamViewersHistoryRepository) {
				rr.countByLivestreamOwnerID = func(context.Context, repository.Querier, model.UserID) (int64, error) { return 0, boom }
			},
			wantErr: boom,
			wantMsg: "failed to count reactions: boom",
		},
		{
			name: "count tips fails",
			modify: func(_ *fakeUserRepository, _ *fakeLivestreamRepository, lr *fakeLivecommentRepository, _ *fakeReactionRepository, _ *fakeLivestreamViewersHistoryRepository) {
				lr.sumTipByLivestreamOwnerID = func(context.Context, repository.Querier, model.UserID) (int64, error) { return 0, boom }
			},
			wantErr: boom,
			wantMsg: "failed to count tips: boom",
		},
		{
			name: "count total reactions fails",
			modify: func(_ *fakeUserRepository, _ *fakeLivestreamRepository, _ *fakeLivecommentRepository, rr *fakeReactionRepository, _ *fakeLivestreamViewersHistoryRepository) {
				rr.countByLivestreamOwnerName = func(context.Context, repository.Querier, string) (int64, error) { return 0, boom }
			},
			wantErr: boom,
			wantMsg: "failed to count total reactions: boom",
		},
		{
			name: "get livestreams fails",
			modify: func(_ *fakeUserRepository, lsr *fakeLivestreamRepository, _ *fakeLivecommentRepository, _ *fakeReactionRepository, _ *fakeLivestreamViewersHistoryRepository) {
				lsr.findAllByUserID = func(context.Context, repository.Querier, model.UserID) ([]*model.LivestreamModel, error) {
					return nil, boom
				}
			},
			wantErr: boom,
			wantMsg: "failed to get livestreams: boom",
		},
		{
			name: "get livecomments fails",
			modify: func(_ *fakeUserRepository, _ *fakeLivestreamRepository, lr *fakeLivecommentRepository, _ *fakeReactionRepository, _ *fakeLivestreamViewersHistoryRepository) {
				lr.findAllByLivestreamID = func(context.Context, repository.Querier, model.LivestreamID) ([]*model.LivecommentModel, error) {
					return nil, boom
				}
			},
			wantErr: boom,
			wantMsg: "failed to get livecomments: boom",
		},
		{
			name: "count viewers fails",
			modify: func(_ *fakeUserRepository, _ *fakeLivestreamRepository, _ *fakeLivecommentRepository, _ *fakeReactionRepository, vr *fakeLivestreamViewersHistoryRepository) {
				vr.countByLivestreamID = func(context.Context, repository.Querier, model.LivestreamID) (int64, error) { return 0, boom }
			},
			wantErr: boom,
			wantMsg: "failed to get livestream_view_history: boom",
		},
		{
			name: "find favorite emoji fails",
			modify: func(_ *fakeUserRepository, _ *fakeLivestreamRepository, _ *fakeLivecommentRepository, rr *fakeReactionRepository, _ *fakeLivestreamViewersHistoryRepository) {
				rr.findFavoriteEmojiByLivestreamOwnerName = func(context.Context, repository.Querier, string) (string, error) { return "", boom }
			},
			wantErr: boom,
			wantMsg: "failed to find favorite emoji: boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo, livestreamRepo, livecommentRepo, reactionRepo, viewerRepo := newTestUserStatisticsRepos()
			tt.modify(userRepo, livestreamRepo, livecommentRepo, reactionRepo, viewerRepo)
			u := NewStatisticsUsecase(&fakeTxManager{}, userRepo, livestreamRepo, livecommentRepo, reactionRepo, viewerRepo, &fakeLivecommentReportRepository{})

			_, err := u.FindUserStatistics(context.Background(), "bob")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if tt.wantMsg != "" && err.Error() != tt.wantMsg {
				t.Errorf("err = %q, want %q", err.Error(), tt.wantMsg)
			}
		})
	}
}

func newTestLivestreamStatisticsRepos() (*fakeLivestreamRepository, *fakeLivecommentRepository, *fakeReactionRepository, *fakeLivestreamViewersHistoryRepository, *fakeLivecommentReportRepository) {
	livestreamRepo := &fakeLivestreamRepository{
		findByID: func(_ context.Context, _ repository.Querier, id model.LivestreamID) (*model.LivestreamModel, error) {
			if id != 11 {
				return nil, fmt.Errorf("unexpected livestreamID %d", id)
			}
			return &model.LivestreamModel{ID: 11}, nil
		},
		findAll: func(context.Context, repository.Querier) ([]*model.LivestreamModel, error) {
			return []*model.LivestreamModel{{ID: 10}, {ID: 11}, {ID: 12}}, nil
		},
	}
	// スコアは リアクション数 + チップ合計: 10 → 30, 11 → 30, 12 → 5
	tips := map[model.LivestreamID]int64{10: 20, 11: 25, 12: 5}
	livecommentRepo := &fakeLivecommentRepository{
		sumTipByLivestreamID: func(_ context.Context, _ repository.Querier, livestreamID model.LivestreamID) (int64, error) {
			return tips[livestreamID], nil
		},
		maxTipByLivestreamID: func(_ context.Context, _ repository.Querier, livestreamID model.LivestreamID) (int64, error) {
			if livestreamID != 11 {
				return 0, fmt.Errorf("unexpected livestreamID %d", livestreamID)
			}
			return 20, nil
		},
	}
	reactionCounts := map[model.LivestreamID]int64{10: 10, 11: 5, 12: 0}
	reactionRepo := &fakeReactionRepository{
		countByLivestreamID: func(_ context.Context, _ repository.Querier, livestreamID model.LivestreamID) (int64, error) {
			return reactionCounts[livestreamID], nil
		},
		countTotalByLivestreamID: func(_ context.Context, _ repository.Querier, livestreamID model.LivestreamID) (int64, error) {
			if livestreamID != 11 {
				return 0, fmt.Errorf("unexpected livestreamID %d", livestreamID)
			}
			return 5, nil
		},
	}
	viewerRepo := &fakeLivestreamViewersHistoryRepository{
		countViewersByLivestreamID: func(_ context.Context, _ repository.Querier, livestreamID model.LivestreamID) (int64, error) {
			if livestreamID != 11 {
				return 0, fmt.Errorf("unexpected livestreamID %d", livestreamID)
			}
			return 7, nil
		},
	}
	reportRepo := &fakeLivecommentReportRepository{
		countByLivestreamID: func(_ context.Context, _ repository.Querier, livestreamID model.LivestreamID) (int64, error) {
			if livestreamID != 11 {
				return 0, fmt.Errorf("unexpected livestreamID %d", livestreamID)
			}
			return 2, nil
		},
	}
	return livestreamRepo, livecommentRepo, reactionRepo, viewerRepo, reportRepo
}

func TestStatisticsUsecase_FindLivestreamStatistics(t *testing.T) {
	livestreamRepo, livecommentRepo, reactionRepo, viewerRepo, reportRepo := newTestLivestreamStatisticsRepos()
	// 呼ばれた順を確認するため、準備した関数を記録用の関数で包む
	var livestreamCalls []string
	findByID, findAll := livestreamRepo.findByID, livestreamRepo.findAll
	livestreamRepo.findByID = func(ctx context.Context, q repository.Querier, id model.LivestreamID) (*model.LivestreamModel, error) {
		livestreamCalls = append(livestreamCalls, "FindByID")
		return findByID(ctx, q, id)
	}
	livestreamRepo.findAll = func(ctx context.Context, q repository.Querier) ([]*model.LivestreamModel, error) {
		livestreamCalls = append(livestreamCalls, "FindAll")
		return findAll(ctx, q)
	}
	u := NewStatisticsUsecase(&fakeTxManager{}, &fakeUserRepository{}, livestreamRepo, livecommentRepo, reactionRepo, viewerRepo, reportRepo)

	got, err := u.FindLivestreamStatistics(context.Background(), 11)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := model.LivestreamStatistics{
		// ID 10 と同点だが、ID の昇順で後ろの 11 が上位になる
		Rank:           1,
		ViewersCount:   7,
		TotalReactions: 5,
		TotalReports:   2,
		MaxTip:         20,
	}
	if *got != want {
		t.Errorf("got %+v, want %+v", *got, want)
	}
	// 対象のライブ配信の存在を確認してから、全ライブ配信を取得する
	if want := []string{"FindByID", "FindAll"}; !slices.Equal(livestreamCalls, want) {
		t.Errorf("livestream calls = %v, want %v", livestreamCalls, want)
	}
}

func TestStatisticsUsecase_FindLivestreamStatistics_Errors(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name    string
		modify  func(*fakeLivestreamRepository, *fakeLivecommentRepository, *fakeReactionRepository, *fakeLivestreamViewersHistoryRepository, *fakeLivecommentReportRepository)
		wantErr error
		wantMsg string
	}{
		{
			name: "livestream not found",
			modify: func(lsr *fakeLivestreamRepository, _ *fakeLivecommentRepository, _ *fakeReactionRepository, _ *fakeLivestreamViewersHistoryRepository, _ *fakeLivecommentReportRepository) {
				lsr.findByID = func(context.Context, repository.Querier, model.LivestreamID) (*model.LivestreamModel, error) {
					return nil, repository.ErrNotFound
				}
			},
			wantErr: ErrLivestreamNotFound,
		},
		{
			name: "get livestream fails",
			modify: func(lsr *fakeLivestreamRepository, _ *fakeLivecommentRepository, _ *fakeReactionRepository, _ *fakeLivestreamViewersHistoryRepository, _ *fakeLivecommentReportRepository) {
				lsr.findByID = func(context.Context, repository.Querier, model.LivestreamID) (*model.LivestreamModel, error) {
					return nil, boom
				}
			},
			wantErr: boom,
			wantMsg: "failed to get livestream: boom",
		},
		{
			name: "get livestreams fails",
			modify: func(lsr *fakeLivestreamRepository, _ *fakeLivecommentRepository, _ *fakeReactionRepository, _ *fakeLivestreamViewersHistoryRepository, _ *fakeLivecommentReportRepository) {
				lsr.findAll = func(context.Context, repository.Querier) ([]*model.LivestreamModel, error) { return nil, boom }
			},
			wantErr: boom,
			wantMsg: "failed to get livestreams: boom",
		},
		{
			name: "count reactions fails",
			modify: func(_ *fakeLivestreamRepository, _ *fakeLivecommentRepository, rr *fakeReactionRepository, _ *fakeLivestreamViewersHistoryRepository, _ *fakeLivecommentReportRepository) {
				rr.countByLivestreamID = func(context.Context, repository.Querier, model.LivestreamID) (int64, error) { return 0, boom }
			},
			wantErr: boom,
			wantMsg: "failed to count reactions: boom",
		},
		{
			name: "count tips fails",
			modify: func(_ *fakeLivestreamRepository, lr *fakeLivecommentRepository, _ *fakeReactionRepository, _ *fakeLivestreamViewersHistoryRepository, _ *fakeLivecommentReportRepository) {
				lr.sumTipByLivestreamID = func(context.Context, repository.Querier, model.LivestreamID) (int64, error) { return 0, boom }
			},
			wantErr: boom,
			wantMsg: "failed to count tips: boom",
		},
		{
			name: "count viewers fails",
			modify: func(_ *fakeLivestreamRepository, _ *fakeLivecommentRepository, _ *fakeReactionRepository, vr *fakeLivestreamViewersHistoryRepository, _ *fakeLivecommentReportRepository) {
				vr.countViewersByLivestreamID = func(context.Context, repository.Querier, model.LivestreamID) (int64, error) { return 0, boom }
			},
			wantErr: boom,
			wantMsg: "failed to count livestream viewers: boom",
		},
		{
			name: "max tip fails",
			modify: func(_ *fakeLivestreamRepository, lr *fakeLivecommentRepository, _ *fakeReactionRepository, _ *fakeLivestreamViewersHistoryRepository, _ *fakeLivecommentReportRepository) {
				lr.maxTipByLivestreamID = func(context.Context, repository.Querier, model.LivestreamID) (int64, error) { return 0, boom }
			},
			wantErr: boom,
			wantMsg: "failed to find maximum tip livecomment: boom",
		},
		{
			name: "count total reactions fails",
			modify: func(_ *fakeLivestreamRepository, _ *fakeLivecommentRepository, rr *fakeReactionRepository, _ *fakeLivestreamViewersHistoryRepository, _ *fakeLivecommentReportRepository) {
				rr.countTotalByLivestreamID = func(context.Context, repository.Querier, model.LivestreamID) (int64, error) { return 0, boom }
			},
			wantErr: boom,
			wantMsg: "failed to count total reactions: boom",
		},
		{
			name: "count reports fails",
			modify: func(_ *fakeLivestreamRepository, _ *fakeLivecommentRepository, _ *fakeReactionRepository, _ *fakeLivestreamViewersHistoryRepository, rpr *fakeLivecommentReportRepository) {
				rpr.countByLivestreamID = func(context.Context, repository.Querier, model.LivestreamID) (int64, error) { return 0, boom }
			},
			wantErr: boom,
			wantMsg: "failed to count total spam reports: boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			livestreamRepo, livecommentRepo, reactionRepo, viewerRepo, reportRepo := newTestLivestreamStatisticsRepos()
			tt.modify(livestreamRepo, livecommentRepo, reactionRepo, viewerRepo, reportRepo)
			u := NewStatisticsUsecase(&fakeTxManager{}, &fakeUserRepository{}, livestreamRepo, livecommentRepo, reactionRepo, viewerRepo, reportRepo)

			_, err := u.FindLivestreamStatistics(context.Background(), 11)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if tt.wantMsg != "" && err.Error() != tt.wantMsg {
				t.Errorf("err = %q, want %q", err.Error(), tt.wantMsg)
			}
		})
	}
}
