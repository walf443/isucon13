package usecase

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

func TestLivestreamFiller_Fill(t *testing.T) {
	f := testDetailFixture()

	// 同じライブ配信・配信者・タグが重複していても 1 回だけ取得する
	got, err := f.livestreamFiller().Fill(context.Background(), nil, []*domain.Livestream{testLivestream1, testLivestream2, testLivestream1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := map[domain.LivestreamID]*domain.LivestreamDetail{
		1: f.livestream(testLivestream1),
		2: f.livestream(testLivestream2),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("livestreams = %+v, want %+v", got, want)
	}
	if want := []string{"users [42]", "livestream tags [1 2]", "tags [7 8]"}; !reflect.DeepEqual(f.calls, want) {
		t.Errorf("calls = %v, want %v", f.calls, want)
	}
}

func TestLivestreamFiller_Fill_Errors(t *testing.T) {
	tests := []struct {
		name string
		// modify はデータを欠けさせる
		modify  func(f *detailFixture)
		wantErr error
		wantMsg string
	}{
		{
			name:    "owner not found",
			modify:  func(f *detailFixture) { delete(f.users, 42) },
			wantErr: errMissingDetail,
			wantMsg: "failed to get owner of livestream 1: not found",
		},
		{
			name:    "tag not found",
			modify:  func(f *detailFixture) { delete(f.tags, 8) },
			wantErr: errMissingDetail,
			wantMsg: "failed to get tag 8 of livestream 1: not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := testDetailFixture()
			tt.modify(f)
			_, err := f.livestreamFiller().Fill(context.Background(), nil, []*domain.Livestream{testLivestream1})
			if !errors.Is(err, tt.wantErr) || (tt.wantErr == errMissingDetail && errors.Is(err, repository.ErrNotFound)) || err.Error() != tt.wantMsg {
				t.Errorf("err = %v, want %q", err, tt.wantMsg)
			}
		})
	}
}

func TestLivestreamFiller_Fill_RepositoryErrors(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name string
		// modify は filler の repository を失敗するものに差し替える
		modify  func(filler *LivestreamFiller)
		wantMsg string
	}{
		{
			name: "get owners fails",
			modify: func(filler *LivestreamFiller) {
				filler.userRepo = &fakeUserRepository{
					findAllByIDs: func(context.Context, repository.Querier, []domain.UserID) ([]*domain.User, error) { return nil, boom },
				}
			},
			wantMsg: "failed to get owners: boom",
		},
		{
			name: "get livestream tags fails",
			modify: func(filler *LivestreamFiller) {
				filler.livestreamTagRepo = &fakeLivestreamTagRepository{
					findAllByLivestreamIDs: func(context.Context, repository.Querier, []domain.LivestreamID) ([]*domain.LivestreamTag, error) {
						return nil, boom
					},
				}
			},
			wantMsg: "failed to get livestream tags: boom",
		},
		{
			name: "get tags fails",
			modify: func(filler *LivestreamFiller) {
				filler.tagRepo = &fakeTagRepository{
					findAllByIDs: func(context.Context, repository.Querier, []domain.TagID) ([]*domain.Tag, error) { return nil, boom },
				}
			},
			wantMsg: "failed to get tags: boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filler := testDetailFixture().livestreamFiller()
			tt.modify(filler)
			_, err := filler.Fill(context.Background(), nil, []*domain.Livestream{testLivestream1})
			if !errors.Is(err, boom) || err.Error() != tt.wantMsg {
				t.Errorf("err = %v, want %q", err, tt.wantMsg)
			}
		})
	}
}
