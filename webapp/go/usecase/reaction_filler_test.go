package usecase

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

func TestReactionFiller_Fill(t *testing.T) {
	f := testDetailFixture()

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
	if want := []string{"user 43", "user 42", "users [42]", "livestream tags [1 2]", "tags [7 8]"}; !reflect.DeepEqual(f.calls, want) {
		t.Errorf("calls = %v, want %v", f.calls, want)
	}
}

func TestReactionFiller_Fill_Errors(t *testing.T) {
	tests := []struct {
		name        string
		modify      func(f *detailFixture)
		livestreams []*domain.Livestream
		wantMsg     string
	}{
		{
			name:        "user not found",
			modify:      func(f *detailFixture) { delete(f.users, 43) },
			livestreams: []*domain.Livestream{testLivestream1},
			wantMsg:     "failed to get user of reaction 1: not found",
		},
		{
			name:    "livestream not found",
			modify:  func(*detailFixture) {},
			wantMsg: "failed to get livestream of reaction 1: not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := testDetailFixture()
			tt.modify(f)
			_, err := f.reactionFiller(tt.livestreams...).Fill(context.Background(), nil, []*domain.Reaction{testReaction1})
			if !errors.Is(err, errMissingDetail) || errors.Is(err, repository.ErrNotFound) || err.Error() != tt.wantMsg {
				t.Errorf("err = %v, want %q", err, tt.wantMsg)
			}
		})
	}
}
