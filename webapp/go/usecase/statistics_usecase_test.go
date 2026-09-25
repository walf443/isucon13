package usecase

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

func newTestUserStatisticsRepos() (*fakeUserRepository, *fakeLivestreamRepository, *fakeLivecommentRepository, *fakeReactionRepository, *fakeLivestreamViewersHistoryRepository) {
	userRepo := &fakeUserRepository{
		user: &model.UserModel{ID: 2, Name: "bob"},
		users: []*model.UserModel{
			{ID: 1, Name: "alice"},
			{ID: 2, Name: "bob"},
			{ID: 3, Name: "carol"},
		},
	}
	livestreamRepo := &fakeLivestreamRepository{livestreamModels: []*model.LivestreamModel{{ID: 10, UserID: 2}, {ID: 11, UserID: 2}}}
	livecommentRepo := &fakeLivecommentRepository{
		// スコアは リアクション数 + チップ合計: alice 30, bob 30, carol 5
		tipsByOwnerID: map[model.UserID]int64{1: 20, 2: 25, 3: 5},
		livecommentModelsByLivestreamID: map[model.LivestreamID][]*model.LivecommentModel{
			10: {{Tip: 10}, {Tip: 0}},
			11: {{Tip: 15}},
		},
	}
	reactionRepo := &fakeReactionRepository{
		countsByOwnerID:  map[model.UserID]int64{1: 10, 2: 5, 3: 0},
		countByOwnerName: 5,
		favoriteEmoji:    "smile",
	}
	viewerRepo := &fakeLivestreamViewersHistoryRepository{countsByLivestreamID: map[model.LivestreamID]int64{10: 3, 11: 4}}
	return userRepo, livestreamRepo, livecommentRepo, reactionRepo, viewerRepo
}

func TestStatisticsUsecase_FindUserStatistics(t *testing.T) {
	userRepo, livestreamRepo, livecommentRepo, reactionRepo, viewerRepo := newTestUserStatisticsRepos()
	u := NewStatisticsUsecase(&fakeTxManager{}, userRepo, livestreamRepo, livecommentRepo, reactionRepo, viewerRepo)

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
	if userRepo.gotName != "bob" || livestreamRepo.gotUserID != 2 {
		t.Errorf("user name = %q, livestream owner = %d", userRepo.gotName, livestreamRepo.gotUserID)
	}
	if want := []string{"bob", "bob"}; !slices.Equal(reactionRepo.gotOwnerNames, want) {
		t.Errorf("owner names = %v, want %v", reactionRepo.gotOwnerNames, want)
	}
}

func TestStatisticsUsecase_FindUserStatistics_NoReactions(t *testing.T) {
	userRepo, livestreamRepo, livecommentRepo, reactionRepo, viewerRepo := newTestUserStatisticsRepos()
	reactionRepo.favoriteEmoji = ""
	reactionRepo.favoriteEmojiErr = repository.ErrNotFound
	u := NewStatisticsUsecase(&fakeTxManager{}, userRepo, livestreamRepo, livecommentRepo, reactionRepo, viewerRepo)

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
				ur.err = repository.ErrNotFound
			},
			wantErr: ErrUserNotFound,
		},
		{
			name: "get user fails",
			modify: func(ur *fakeUserRepository, _ *fakeLivestreamRepository, _ *fakeLivecommentRepository, _ *fakeReactionRepository, _ *fakeLivestreamViewersHistoryRepository) {
				ur.err = boom
			},
			wantErr: boom,
			wantMsg: "failed to get user: boom",
		},
		{
			name: "get users fails",
			modify: func(ur *fakeUserRepository, _ *fakeLivestreamRepository, _ *fakeLivecommentRepository, _ *fakeReactionRepository, _ *fakeLivestreamViewersHistoryRepository) {
				ur.findAllErr = boom
			},
			wantErr: boom,
			wantMsg: "failed to get users: boom",
		},
		{
			name: "count reactions fails",
			modify: func(_ *fakeUserRepository, _ *fakeLivestreamRepository, _ *fakeLivecommentRepository, rr *fakeReactionRepository, _ *fakeLivestreamViewersHistoryRepository) {
				rr.countByOwnerErr = boom
			},
			wantErr: boom,
			wantMsg: "failed to count reactions: boom",
		},
		{
			name: "count tips fails",
			modify: func(_ *fakeUserRepository, _ *fakeLivestreamRepository, lr *fakeLivecommentRepository, _ *fakeReactionRepository, _ *fakeLivestreamViewersHistoryRepository) {
				lr.sumTipErr = boom
			},
			wantErr: boom,
			wantMsg: "failed to count tips: boom",
		},
		{
			name: "count total reactions fails",
			modify: func(_ *fakeUserRepository, _ *fakeLivestreamRepository, _ *fakeLivecommentRepository, rr *fakeReactionRepository, _ *fakeLivestreamViewersHistoryRepository) {
				rr.countByNameErr = boom
			},
			wantErr: boom,
			wantMsg: "failed to count total reactions: boom",
		},
		{
			name: "get livestreams fails",
			modify: func(_ *fakeUserRepository, lsr *fakeLivestreamRepository, _ *fakeLivecommentRepository, _ *fakeReactionRepository, _ *fakeLivestreamViewersHistoryRepository) {
				lsr.err = boom
			},
			wantErr: boom,
			wantMsg: "failed to get livestreams: boom",
		},
		{
			name: "get livecomments fails",
			modify: func(_ *fakeUserRepository, _ *fakeLivestreamRepository, lr *fakeLivecommentRepository, _ *fakeReactionRepository, _ *fakeLivestreamViewersHistoryRepository) {
				lr.findAllByLivestreamIDErr = boom
			},
			wantErr: boom,
			wantMsg: "failed to get livecomments: boom",
		},
		{
			name: "count viewers fails",
			modify: func(_ *fakeUserRepository, _ *fakeLivestreamRepository, _ *fakeLivecommentRepository, _ *fakeReactionRepository, vr *fakeLivestreamViewersHistoryRepository) {
				vr.countErr = boom
			},
			wantErr: boom,
			wantMsg: "failed to get livestream_view_history: boom",
		},
		{
			name: "find favorite emoji fails",
			modify: func(_ *fakeUserRepository, _ *fakeLivestreamRepository, _ *fakeLivecommentRepository, rr *fakeReactionRepository, _ *fakeLivestreamViewersHistoryRepository) {
				rr.favoriteEmojiErr = boom
			},
			wantErr: boom,
			wantMsg: "failed to find favorite emoji: boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo, livestreamRepo, livecommentRepo, reactionRepo, viewerRepo := newTestUserStatisticsRepos()
			tt.modify(userRepo, livestreamRepo, livecommentRepo, reactionRepo, viewerRepo)
			u := NewStatisticsUsecase(&fakeTxManager{}, userRepo, livestreamRepo, livecommentRepo, reactionRepo, viewerRepo)

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
