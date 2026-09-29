package usecase

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

func TestLivecommentFiller_Fill(t *testing.T) {
	f := testDetailFixture()

	got, err := f.livecommentFiller(testLivestream1, testLivestream2).Fill(context.Background(), nil, []*domain.Livecomment{testLivecomment1, testLivecomment2, testLivecomment3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := map[domain.LivecommentID]*domain.LivecommentDetail{
		50: f.livecomment(testLivecomment1, testLivestream1),
		51: f.livecomment(testLivecomment2, testLivestream1),
		52: f.livecomment(testLivecomment3, testLivestream2),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("livecomments = %+v, want %+v", got, want)
	}
	// コメントしたユーザは重複を除いて 1 回ずつ引く (配信者は LivestreamFiller が別に引く)
	if want := []string{"user 43", "user 42", "user 42", "livestream tags 1", "tag 7", "tag 8", "livestream tags 2"}; !reflect.DeepEqual(f.calls, want) {
		t.Errorf("calls = %v, want %v", f.calls, want)
	}
}

func TestLivecommentFiller_Fill_Errors(t *testing.T) {
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
			wantMsg:     "failed to get user of livecomment 50: not found",
		},
		{
			name:    "livestream not found",
			modify:  func(*detailFixture) {},
			wantMsg: "failed to get livestream of livecomment 50: not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := testDetailFixture()
			tt.modify(f)
			_, err := f.livecommentFiller(tt.livestreams...).Fill(context.Background(), nil, []*domain.Livecomment{testLivecomment1})
			if !errors.Is(err, errMissingDetail) || errors.Is(err, repository.ErrNotFound) || err.Error() != tt.wantMsg {
				t.Errorf("err = %v, want %q", err, tt.wantMsg)
			}
		})
	}
}
