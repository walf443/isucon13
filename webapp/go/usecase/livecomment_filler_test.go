package usecase

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// newLivecommentRepositoryWithLivecomments は ID で models を引ける fakeLivecommentRepository を返す。
// 見つからない ID には ErrNotFound を返す。
func newLivecommentRepositoryWithLivecomments(models ...*domain.Livecomment) *fakeLivecommentRepository {
	return &fakeLivecommentRepository{
		findByID: func(_ context.Context, _ repository.Querier, id domain.LivecommentID) (*domain.Livecomment, error) {
			for _, m := range models {
				if m.ID == id {
					return m, nil
				}
			}
			return nil, repository.ErrNotFound
		},
	}
}

// livecommentFiller は f のユーザと livestreams のライブ配信を引く LivecommentFiller を返す。
func (f *livestreamFixture) livecommentFiller(livestreams ...*domain.Livestream) *LivecommentFiller {
	return NewLivecommentFiller(f.userRepo(), newLivestreamRepositoryWithLivestreams(nil, livestreams...), f.userFiller(), f.filler())
}

// livecomment は、このデータで livecomment を埋めた結果として期待する domain.LivecommentDetail を返す。
// livestream は livecomment のライブ配信。
func (f *livestreamFixture) livecomment(livecomment *domain.Livecomment, livestream *domain.Livestream) *domain.LivecommentDetail {
	return &domain.LivecommentDetail{
		ID:         livecomment.ID,
		User:       f.user(livecomment.UserID),
		Livestream: *f.livestream(livestream),
		Comment:    livecomment.Comment,
		Tip:        livecomment.Tip,
		CreatedAt:  livecomment.CreatedAt,
	}
}

var (
	testLivecomment1 = &domain.Livecomment{ID: 50, UserID: 43, LivestreamID: 1, Comment: "hello", Tip: 10, CreatedAt: 300}
	testLivecomment2 = &domain.Livecomment{ID: 51, UserID: 42, LivestreamID: 1, Comment: "thanks", CreatedAt: 200}
	testLivecomment3 = &domain.Livecomment{ID: 52, UserID: 43, LivestreamID: 2, Comment: "hi", CreatedAt: 100}
)

func TestLivecommentFiller_Fill(t *testing.T) {
	f := testLivestreamFixture()

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
		modify      func(f *livestreamFixture)
		livestreams []*domain.Livestream
		wantMsg     string
	}{
		{
			name:        "user not found",
			modify:      func(f *livestreamFixture) { delete(f.users, 43) },
			livestreams: []*domain.Livestream{testLivestream1},
			wantMsg:     "failed to get user of livecomment 50: not found",
		},
		{
			name:    "livestream not found",
			modify:  func(*livestreamFixture) {},
			wantMsg: "failed to get livestream of livecomment 50: not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := testLivestreamFixture()
			tt.modify(f)
			_, err := f.livecommentFiller(tt.livestreams...).Fill(context.Background(), nil, []*domain.Livecomment{testLivecomment1})
			if !errors.Is(err, errMissingDetail) || errors.Is(err, repository.ErrNotFound) || err.Error() != tt.wantMsg {
				t.Errorf("err = %v, want %q", err, tt.wantMsg)
			}
		})
	}
}
