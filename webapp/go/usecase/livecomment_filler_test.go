package usecase

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// newLivecommentRepositoryWithModels は ID で models を引ける fakeLivecommentRepository を返す。
// 見つからない ID には ErrNotFound を返す。
func newLivecommentRepositoryWithModels(models ...*domain.LivecommentModel) *fakeLivecommentRepository {
	return &fakeLivecommentRepository{
		findByID: func(_ context.Context, _ repository.Querier, id domain.LivecommentID) (*domain.LivecommentModel, error) {
			for _, m := range models {
				if m.ID == id {
					return m, nil
				}
			}
			return nil, repository.ErrNotFound
		},
	}
}

// livecommentFiller は f のユーザと livestreamModels のライブ配信を引く LivecommentFiller を返す。
func (f *livestreamFixture) livecommentFiller(livestreamModels ...*domain.LivestreamModel) *LivecommentFiller {
	return NewLivecommentFiller(f.userRepo(), newLivestreamRepositoryWithModels(nil, livestreamModels...), f.userFiller(), f.filler())
}

// livecomment は、このデータで livecommentModel を埋めた結果として期待する domain.Livecomment を返す。
// livestreamModel は livecommentModel のライブ配信。
func (f *livestreamFixture) livecomment(livecommentModel *domain.LivecommentModel, livestreamModel *domain.LivestreamModel) *domain.LivecommentDetail {
	return &domain.LivecommentDetail{
		ID:         livecommentModel.ID,
		User:       f.user(livecommentModel.UserID),
		Livestream: *f.livestream(livestreamModel),
		Comment:    livecommentModel.Comment,
		Tip:        livecommentModel.Tip,
		CreatedAt:  livecommentModel.CreatedAt,
	}
}

var (
	testLivecommentModel1 = &domain.LivecommentModel{ID: 50, UserID: 43, LivestreamID: 1, Comment: "hello", Tip: 10, CreatedAt: 300}
	testLivecommentModel2 = &domain.LivecommentModel{ID: 51, UserID: 42, LivestreamID: 1, Comment: "thanks", CreatedAt: 200}
	testLivecommentModel3 = &domain.LivecommentModel{ID: 52, UserID: 43, LivestreamID: 2, Comment: "hi", CreatedAt: 100}
)

func TestLivecommentFiller_Fill(t *testing.T) {
	f := testLivestreamFixture()

	got, err := f.livecommentFiller(testLivestreamModel1, testLivestreamModel2).Fill(context.Background(), nil, []*domain.LivecommentModel{testLivecommentModel1, testLivecommentModel2, testLivecommentModel3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := map[domain.LivecommentID]*domain.LivecommentDetail{
		50: f.livecomment(testLivecommentModel1, testLivestreamModel1),
		51: f.livecomment(testLivecommentModel2, testLivestreamModel1),
		52: f.livecomment(testLivecommentModel3, testLivestreamModel2),
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
		name             string
		modify           func(f *livestreamFixture)
		livestreamModels []*domain.LivestreamModel
		wantMsg          string
	}{
		{
			name:             "user not found",
			modify:           func(f *livestreamFixture) { delete(f.users, 43) },
			livestreamModels: []*domain.LivestreamModel{testLivestreamModel1},
			wantMsg:          "failed to get user of livecomment 50: not found",
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
			_, err := f.livecommentFiller(tt.livestreamModels...).Fill(context.Background(), nil, []*domain.LivecommentModel{testLivecommentModel1})
			if !errors.Is(err, repository.ErrNotFound) || err.Error() != tt.wantMsg {
				t.Errorf("err = %v, want %q", err, tt.wantMsg)
			}
		})
	}
}
