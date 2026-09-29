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
	if want := []string{"user 42", "livestream tags 1", "tag 7", "tag 8", "livestream tags 2"}; !reflect.DeepEqual(f.calls, want) {
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
	f := testDetailFixture()
	filler := f.livestreamFiller()
	filler.livestreamTagRepo = &fakeLivestreamTagRepository{
		findAllByLivestreamID: func(context.Context, repository.Querier, domain.LivestreamID) ([]*domain.LivestreamTag, error) {
			return nil, boom
		},
	}

	_, err := filler.Fill(context.Background(), nil, []*domain.Livestream{testLivestream1})
	if want := "failed to get tags of livestream 1: boom"; !errors.Is(err, boom) || err.Error() != want {
		t.Errorf("err = %v, want %q", err, want)
	}
}
