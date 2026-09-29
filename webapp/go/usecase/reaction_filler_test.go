package usecase

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// reactionFiller は f のユーザと livestreams のライブ配信を引く ReactionFiller を返す。
func (f *livestreamFixture) reactionFiller(livestreams ...*domain.Livestream) *ReactionFiller {
	return NewReactionFiller(f.userRepo(), newLivestreamRepositoryWithLivestreams(nil, livestreams...), f.userFiller(), f.filler())
}

// reaction は、このデータで reaction を埋めた結果として期待する domain.ReactionDetail を返す。
// livestream は reaction のライブ配信。
func (f *livestreamFixture) reaction(reaction *domain.Reaction, livestream *domain.Livestream) *domain.ReactionDetail {
	return &domain.ReactionDetail{
		ID:         reaction.ID,
		EmojiName:  reaction.EmojiName,
		User:       f.user(reaction.UserID),
		Livestream: *f.livestream(livestream),
		CreatedAt:  reaction.CreatedAt,
	}
}

var (
	testReaction1 = &domain.Reaction{ID: 1, UserID: 43, LivestreamID: 1, EmojiName: "tada", CreatedAt: 300}
	testReaction2 = &domain.Reaction{ID: 2, UserID: 42, LivestreamID: 1, EmojiName: "heart", CreatedAt: 200}
	testReaction3 = &domain.Reaction{ID: 3, UserID: 43, LivestreamID: 2, EmojiName: "tada", CreatedAt: 100}
)

func TestReactionFiller_Fill(t *testing.T) {
	f := testLivestreamFixture()

	got, err := f.reactionFiller(testLivestream1, testLivestream2).Fill(context.Background(), nil, []*domain.Reaction{testReaction1, testReaction2, testReaction3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := map[domain.ReactionID]*domain.ReactionDetail{
		1: f.reaction(testReaction1, testLivestream1),
		2: f.reaction(testReaction2, testLivestream1),
		3: f.reaction(testReaction3, testLivestream2),
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
		name        string
		modify      func(f *livestreamFixture)
		livestreams []*domain.Livestream
		wantMsg     string
	}{
		{
			name:        "user not found",
			modify:      func(f *livestreamFixture) { delete(f.users, 43) },
			livestreams: []*domain.Livestream{testLivestream1},
			wantMsg:     "failed to get user of reaction 1: not found",
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
			_, err := f.reactionFiller(tt.livestreams...).Fill(context.Background(), nil, []*domain.Reaction{testReaction1})
			if !errors.Is(err, errMissingDetail) || errors.Is(err, repository.ErrNotFound) || err.Error() != tt.wantMsg {
				t.Errorf("err = %v, want %q", err, tt.wantMsg)
			}
		})
	}
}
