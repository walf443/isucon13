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
	// コメントしたユーザは重複を除いた一覧で 1 回だけ引く (配信者は LivestreamFiller が別に引く)
	if want := []string{"users [43 42]", "users [42]", "livestream tags [1 2]", "tags [7 8]"}; !reflect.DeepEqual(f.calls, want) {
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

func TestLivecommentFiller_Fill_RepositoryErrors(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name    string
		modify  func(filler *LivecommentFiller)
		wantMsg string
	}{
		{
			name: "get users fails",
			modify: func(filler *LivecommentFiller) {
				filler.userRepo = &fakeUserRepository{
					findAllByIDs: func(context.Context, repository.Querier, []domain.UserID) ([]*domain.User, error) { return nil, boom },
				}
			},
			wantMsg: "failed to get users: boom",
		},
		{
			name: "get livestreams fails",
			modify: func(filler *LivecommentFiller) {
				filler.livestreamRepo = &fakeLivestreamRepository{
					findAllByIDs: func(context.Context, repository.Querier, []domain.LivestreamID) ([]*domain.Livestream, error) {
						return nil, boom
					},
				}
			},
			wantMsg: "failed to get livestreams: boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filler := testDetailFixture().livecommentFiller(testLivestream1)
			tt.modify(filler)
			_, err := filler.Fill(context.Background(), nil, []*domain.Livecomment{testLivecomment1})
			if !errors.Is(err, boom) || err.Error() != tt.wantMsg {
				t.Errorf("err = %v, want %q", err, tt.wantMsg)
			}
		})
	}
}
