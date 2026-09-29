package usecase

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// reactionFiller は f のユーザと livestreamModels のライブ配信を引く ReactionFiller を返す。
func (f *livestreamFixture) reactionFiller(livestreamModels ...*domain.Livestream) *ReactionFiller {
	return NewReactionFiller(f.userRepo(), newLivestreamRepositoryWithModels(nil, livestreamModels...), f.userFiller(), f.filler())
}

// reaction は、このデータで reactionModel を埋めた結果として期待する domain.Reaction を返す。
// livestreamModel は reactionModel のライブ配信。
func (f *livestreamFixture) reaction(reactionModel *domain.Reaction, livestreamModel *domain.Livestream) *domain.ReactionDetail {
	return &domain.ReactionDetail{
		ID:         reactionModel.ID,
		EmojiName:  reactionModel.EmojiName,
		User:       f.user(reactionModel.UserID),
		Livestream: *f.livestream(livestreamModel),
		CreatedAt:  reactionModel.CreatedAt,
	}
}

var (
	testReactionModel1 = &domain.Reaction{ID: 1, UserID: 43, LivestreamID: 1, EmojiName: "tada", CreatedAt: 300}
	testReactionModel2 = &domain.Reaction{ID: 2, UserID: 42, LivestreamID: 1, EmojiName: "heart", CreatedAt: 200}
	testReactionModel3 = &domain.Reaction{ID: 3, UserID: 43, LivestreamID: 2, EmojiName: "tada", CreatedAt: 100}
)

func TestReactionFiller_Fill(t *testing.T) {
	f := testLivestreamFixture()

	got, err := f.reactionFiller(testLivestreamModel1, testLivestreamModel2).Fill(context.Background(), nil, []*domain.Reaction{testReactionModel1, testReactionModel2, testReactionModel3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := map[domain.ReactionID]*domain.ReactionDetail{
		1: f.reaction(testReactionModel1, testLivestreamModel1),
		2: f.reaction(testReactionModel2, testLivestreamModel1),
		3: f.reaction(testReactionModel3, testLivestreamModel2),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("reactions = %+v, want %+v", got, want)
	}
	// リアクションしたユーザは重複を除いて 1 回ずつ引く (配信者は LivestreamFiller が別に引く)
	if want := []string{"user 43", "user 42", "user 42", "livestream tags 1", "tag 7", "tag 8", "livestream tags 2"}; !reflect.DeepEqual(f.calls, want) {
		t.Errorf("calls = %v, want %v", f.calls, want)
	}
}

func TestReactionFiller_Fill_Errors(t *testing.T) {
	tests := []struct {
		name             string
		modify           func(f *livestreamFixture)
		livestreamModels []*domain.Livestream
		wantMsg          string
	}{
		{
			name:             "user not found",
			modify:           func(f *livestreamFixture) { delete(f.users, 43) },
			livestreamModels: []*domain.Livestream{testLivestreamModel1},
			wantMsg:          "failed to get user of reaction 1: not found",
		},
		{
			name:    "livestream not found",
			modify:  func(*livestreamFixture) {},
			wantMsg: "failed to get livestream of reaction 1: not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := testLivestreamFixture()
			tt.modify(f)
			_, err := f.reactionFiller(tt.livestreamModels...).Fill(context.Background(), nil, []*domain.Reaction{testReactionModel1})
			if !errors.Is(err, repository.ErrNotFound) || err.Error() != tt.wantMsg {
				t.Errorf("err = %v, want %q", err, tt.wantMsg)
			}
		})
	}
}
