package usecase

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// livestreamFixture は LivestreamFiller などが引くデータ (ユーザ・タグ) を持つ。
// ユーザのテーマは ID がユーザの ID の 10 倍のものとし、アイコンは未登録とする。
type livestreamFixture struct {
	users map[domain.UserID]*domain.UserModel
	tags  map[domain.TagID]*domain.TagModel
	// livestreamTags はライブ配信ごとに付いているタグの ID。
	livestreamTags map[domain.LivestreamID][]domain.TagID
	// calls は呼ばれた repository のメソッドと引数を記録する。
	calls []string
}

// testLivestreamFixture は配信者 alice (ID 42)・視聴者 bob (ID 43) とタグ 7, 8 を持ち、
// ライブ配信 1 にタグ 7, 8、ライブ配信 2 にタグ 7 が付いている。
func testLivestreamFixture() *livestreamFixture {
	return &livestreamFixture{
		users: map[domain.UserID]*domain.UserModel{
			42: {ID: 42, Name: "alice"},
			43: {ID: 43, Name: "bob", DisplayName: "Bob"},
		},
		tags: map[domain.TagID]*domain.TagModel{
			7: {ID: 7, Name: "ゲーム実況"},
			8: {ID: 8, Name: "雑談"},
		},
		livestreamTags: map[domain.LivestreamID][]domain.TagID{1: {7, 8}, 2: {7}},
	}
}

func (f *livestreamFixture) record(call string) {
	f.calls = append(f.calls, call)
}

// userRepo は f のユーザを ID で引ける fakeUserRepository を返す。
func (f *livestreamFixture) userRepo() *fakeUserRepository {
	return &fakeUserRepository{
		findByID: func(_ context.Context, _ repository.Querier, id domain.UserID) (*domain.UserModel, error) {
			f.record(fmt.Sprintf("user %d", id))
			user, ok := f.users[id]
			if !ok {
				return nil, repository.ErrNotFound
			}
			return user, nil
		},
	}
}

// userFiller は f のユーザのテーマを返す UserFiller を返す。
func (f *livestreamFixture) userFiller() *UserFiller {
	themes := map[domain.UserID]*domain.ThemeModel{}
	for id := range f.users {
		themes[id] = &domain.ThemeModel{ID: domain.ThemeID(id * 10), UserID: id}
	}
	return newUserFillerForTest(themes, nil, nil)
}

// user は、このデータでユーザ id を埋めた結果として期待する domain.User を返す。
func (f *livestreamFixture) user(id domain.UserID) domain.UserDetail {
	user := f.users[id]
	return domain.UserDetail{
		ID:          user.ID,
		Name:        user.Name,
		DisplayName: user.DisplayName,
		Description: user.Description,
		Theme:       domain.ThemeModel{ID: domain.ThemeID(user.ID * 10), UserID: user.ID},
		IconHash:    "default-hash",
	}
}

func (f *livestreamFixture) filler() *LivestreamFiller {
	livestreamTagRepo := &fakeLivestreamTagRepository{
		findAllByLivestreamID: func(_ context.Context, _ repository.Querier, livestreamID domain.LivestreamID) ([]*domain.LivestreamTagModel, error) {
			f.record(fmt.Sprintf("livestream tags %d", livestreamID))
			var livestreamTags []*domain.LivestreamTagModel
			for _, tagID := range f.livestreamTags[livestreamID] {
				livestreamTags = append(livestreamTags, &domain.LivestreamTagModel{LivestreamID: livestreamID, TagID: tagID})
			}
			return livestreamTags, nil
		},
	}
	tagRepo := &fakeTagRepository{
		findByID: func(_ context.Context, _ repository.Querier, id domain.TagID) (*domain.TagModel, error) {
			f.record(fmt.Sprintf("tag %d", id))
			tag, ok := f.tags[id]
			if !ok {
				return nil, repository.ErrNotFound
			}
			return tag, nil
		},
	}
	return NewLivestreamFiller(f.userRepo(), livestreamTagRepo, tagRepo, f.userFiller())
}

// livestream は、このデータで livestreamModel を埋めた結果として期待する domain.Livestream を返す。
func (f *livestreamFixture) livestream(livestreamModel *domain.LivestreamModel) *domain.LivestreamDetail {
	tags := make([]domain.TagModel, len(f.livestreamTags[livestreamModel.ID]))
	for i, tagID := range f.livestreamTags[livestreamModel.ID] {
		tags[i] = *f.tags[tagID]
	}
	return &domain.LivestreamDetail{
		ID:           livestreamModel.ID,
		Owner:        f.user(livestreamModel.UserID),
		Title:        livestreamModel.Title,
		Description:  livestreamModel.Description,
		PlaylistUrl:  livestreamModel.PlaylistUrl,
		ThumbnailUrl: livestreamModel.ThumbnailUrl,
		Tags:         tags,
		StartAt:      livestreamModel.StartAt,
		EndAt:        livestreamModel.EndAt,
	}
}

var (
	testLivestreamModel1 = &domain.LivestreamModel{ID: 1, UserID: 42, Title: "first", Description: "desc", PlaylistUrl: "p", ThumbnailUrl: "t", StartAt: 100, EndAt: 200}
	testLivestreamModel2 = &domain.LivestreamModel{ID: 2, UserID: 42, Title: "second"}
)

func TestLivestreamFiller_Fill(t *testing.T) {
	f := testLivestreamFixture()

	// 同じライブ配信・配信者・タグが重複していても 1 回だけ取得する
	got, err := f.filler().Fill(context.Background(), nil, []*domain.LivestreamModel{testLivestreamModel1, testLivestreamModel2, testLivestreamModel1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := map[domain.LivestreamID]*domain.LivestreamDetail{
		1: f.livestream(testLivestreamModel1),
		2: f.livestream(testLivestreamModel2),
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
		modify  func(f *livestreamFixture)
		wantErr error
		wantMsg string
	}{
		{
			name:    "owner not found",
			modify:  func(f *livestreamFixture) { delete(f.users, 42) },
			wantErr: repository.ErrNotFound,
			wantMsg: "failed to get owner of livestream 1: not found",
		},
		{
			name:    "tag not found",
			modify:  func(f *livestreamFixture) { delete(f.tags, 8) },
			wantErr: repository.ErrNotFound,
			wantMsg: "failed to get tag 8 of livestream 1: not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := testLivestreamFixture()
			tt.modify(f)
			_, err := f.filler().Fill(context.Background(), nil, []*domain.LivestreamModel{testLivestreamModel1})
			if !errors.Is(err, tt.wantErr) || err.Error() != tt.wantMsg {
				t.Errorf("err = %v, want %q", err, tt.wantMsg)
			}
		})
	}
}

func TestLivestreamFiller_Fill_RepositoryErrors(t *testing.T) {
	boom := errors.New("boom")
	f := testLivestreamFixture()
	filler := f.filler()
	filler.livestreamTagRepo = &fakeLivestreamTagRepository{
		findAllByLivestreamID: func(context.Context, repository.Querier, domain.LivestreamID) ([]*domain.LivestreamTagModel, error) {
			return nil, boom
		},
	}

	_, err := filler.Fill(context.Background(), nil, []*domain.LivestreamModel{testLivestreamModel1})
	if want := "failed to get tags of livestream 1: boom"; !errors.Is(err, boom) || err.Error() != want {
		t.Errorf("err = %v, want %q", err, want)
	}
}
