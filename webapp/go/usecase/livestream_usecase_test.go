package usecase

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// newLivestreamRepositoryWithModels は ID で models を引ける fakeLivestreamRepository を返す。
// 見つからない ID には ErrNotFound を返し、引いた ID を calls に記録する。
func newLivestreamRepositoryWithModels(calls *[]domain.LivestreamID, models ...*domain.Livestream) *fakeLivestreamRepository {
	return &fakeLivestreamRepository{
		findByID: func(_ context.Context, _ repository.Querier, id domain.LivestreamID) (*domain.Livestream, error) {
			if calls != nil {
				*calls = append(*calls, id)
			}
			for _, m := range models {
				if m.ID == id {
					return m, nil
				}
			}
			return nil, repository.ErrNotFound
		},
	}
}

// newLivestreamUsecaseForTest は f のデータで組み立てる LivestreamUsecase を返す。
func newLivestreamUsecaseForTest(f *livestreamFixture, userRepo *fakeUserRepository, tagRepo *fakeTagRepository, livestreamRepo *fakeLivestreamRepository, livestreamTagRepo *fakeLivestreamTagRepository) LivestreamUsecase {
	return NewLivestreamUsecase(&fakeTxManager{}, userRepo, tagRepo, livestreamRepo, livestreamTagRepo, f.filler())
}

func TestLivestreamUsecase_FindByID(t *testing.T) {
	f := testLivestreamFixture()
	u := newLivestreamUsecaseForTest(f, nil, nil, newLivestreamRepositoryWithModels(nil, testLivestreamModel1), nil)

	got, err := u.FindByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := f.livestream(testLivestreamModel1); !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestLivestreamUsecase_FindByID_Errors(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name           string
		livestreamRepo *fakeLivestreamRepository
		// modify はデータを欠けさせる
		modify  func(f *livestreamFixture)
		wantErr error
		wantMsg string
	}{
		{
			name:           "livestream not found",
			livestreamRepo: newLivestreamRepositoryWithModels(nil),
			wantErr:        ErrLivestreamNotFound,
			wantMsg:        ErrLivestreamNotFound.Error(),
		},
		{
			name: "get livestream fails",
			livestreamRepo: &fakeLivestreamRepository{
				findByID: func(context.Context, repository.Querier, domain.LivestreamID) (*domain.Livestream, error) {
					return nil, boom
				},
			},
			wantErr: boom,
			wantMsg: "failed to get livestream: boom",
		},
		{
			// 配信者の欠損はデータ不整合なので 404 (ErrLivestreamNotFound) にしない
			name:           "owner not found",
			livestreamRepo: newLivestreamRepositoryWithModels(nil, testLivestreamModel1),
			modify:         func(f *livestreamFixture) { delete(f.users, 42) },
			wantErr:        repository.ErrNotFound,
			wantMsg:        "failed to get livestream: failed to get owner of livestream 1: not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := testLivestreamFixture()
			if tt.modify != nil {
				tt.modify(f)
			}
			u := newLivestreamUsecaseForTest(f, nil, nil, tt.livestreamRepo, nil)
			_, err := u.FindByID(context.Background(), 1)
			if !errors.Is(err, tt.wantErr) || err.Error() != tt.wantMsg {
				t.Errorf("err = %v, want %q", err, tt.wantMsg)
			}
			if tt.wantErr != ErrLivestreamNotFound && errors.Is(err, ErrLivestreamNotFound) {
				t.Errorf("err = %v, should not be ErrLivestreamNotFound", err)
			}
		})
	}
}

// newLivestreamRepositoryFindingAllByUserID は配信者 userID のライブ配信として models (失敗させる場合は err) を返す fakeLivestreamRepository を返す。
func newLivestreamRepositoryFindingAllByUserID(t *testing.T, userID domain.UserID, models []*domain.Livestream, err error) *fakeLivestreamRepository {
	return &fakeLivestreamRepository{
		findAllByUserID: func(_ context.Context, _ repository.Querier, gotUserID domain.UserID) ([]*domain.Livestream, error) {
			if gotUserID != userID {
				t.Errorf("userID = %d, want %d", gotUserID, userID)
			}
			return models, err
		},
	}
}

func TestLivestreamUsecase_FindAllByUserID(t *testing.T) {
	f := testLivestreamFixture()
	livestreamRepo := newLivestreamRepositoryFindingAllByUserID(t, 42, []*domain.Livestream{testLivestreamModel2, testLivestreamModel1}, nil)
	u := newLivestreamUsecaseForTest(f, nil, nil, livestreamRepo, nil)

	got, err := u.FindAllByUserID(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// repository が返した順序のまま
	if want := []*domain.LivestreamDetail{f.livestream(testLivestreamModel2), f.livestream(testLivestreamModel1)}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestLivestreamUsecase_FindAllByUserID_Empty(t *testing.T) {
	livestreamRepo := newLivestreamRepositoryFindingAllByUserID(t, 42, nil, nil)
	u := newLivestreamUsecaseForTest(testLivestreamFixture(), nil, nil, livestreamRepo, nil)

	got, err := u.FindAllByUserID(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// JSON で null ではなく [] にするため、空でも nil にしない
	if got == nil || len(got) != 0 {
		t.Errorf("got %#v, want empty non-nil slice", got)
	}
}

func TestLivestreamUsecase_FindAllByUserID_Error(t *testing.T) {
	boom := errors.New("boom")
	livestreamRepo := newLivestreamRepositoryFindingAllByUserID(t, 42, nil, boom)
	u := newLivestreamUsecaseForTest(testLivestreamFixture(), nil, nil, livestreamRepo, nil)

	_, err := u.FindAllByUserID(context.Background(), 42)
	if want := "failed to get livestreams: boom"; !errors.Is(err, boom) || err.Error() != want {
		t.Errorf("err = %v, want %q", err, want)
	}
}

func TestLivestreamUsecase_FindAllByUsername(t *testing.T) {
	f := testLivestreamFixture()
	// ユーザ名から引いた ID で検索する
	livestreamRepo := newLivestreamRepositoryFindingAllByUserID(t, 42, []*domain.Livestream{testLivestreamModel1}, nil)
	u := newLivestreamUsecaseForTest(f, newUserRepositoryFindingID(t, "alice", 42, nil), nil, livestreamRepo, nil)

	got, err := u.FindAllByUsername(context.Background(), "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := []*domain.LivestreamDetail{f.livestream(testLivestreamModel1)}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestLivestreamUsecase_FindAllByUsername_Errors(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name string
		// userID, userErr はユーザ名から ID を引いた結果
		userID  domain.UserID
		userErr error
		// livestreamsErr はライブ配信の検索が返すエラー
		livestreamsErr error
		wantErr        error
		wantMsg        string
	}{
		{name: "user not found", userErr: repository.ErrNotFound, wantErr: ErrUserNotFound, wantMsg: ErrUserNotFound.Error()},
		{name: "user repository error", userErr: boom, wantErr: boom, wantMsg: "failed to get user: boom"},
		{name: "livestream repository error", userID: 42, livestreamsErr: boom, wantErr: boom, wantMsg: "failed to get livestreams: boom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			livestreamRepo := newLivestreamRepositoryFindingAllByUserID(t, 42, nil, tt.livestreamsErr)
			u := newLivestreamUsecaseForTest(testLivestreamFixture(), newUserRepositoryFindingID(t, "alice", tt.userID, tt.userErr), nil, livestreamRepo, nil)
			_, err := u.FindAllByUsername(context.Background(), "alice")
			if !errors.Is(err, tt.wantErr) || err.Error() != tt.wantMsg {
				t.Errorf("err = %v, want %q", err, tt.wantMsg)
			}
		})
	}
}

// newTagRepositoryFindingIDs はタグ名 "ゲーム実況" のタグの ID として tagIDs (失敗させる場合は err) を返す fakeTagRepository を返す。
func newTagRepositoryFindingIDs(t *testing.T, tagIDs []domain.TagID, err error) *fakeTagRepository {
	return &fakeTagRepository{
		findIDsByName: func(_ context.Context, _ repository.Querier, name string) ([]domain.TagID, error) {
			if name != "ゲーム実況" {
				t.Errorf("tag name = %q", name)
			}
			return tagIDs, err
		},
	}
}

// newLivestreamTagRepositoryFindingByTagIDs はタグ 7, 8 の紐付けとして livestreamIDs のライブ配信の紐付け (失敗させる場合は err) を返す。
// 呼ばれた回数を calls に数える。
func newLivestreamTagRepositoryFindingByTagIDs(t *testing.T, calls *int, livestreamIDs []domain.LivestreamID, err error) *fakeLivestreamTagRepository {
	return &fakeLivestreamTagRepository{
		findAllByTagIDs: func(_ context.Context, _ repository.Querier, tagIDs []domain.TagID) ([]*domain.LivestreamTag, error) {
			*calls++
			if !slices.Equal(tagIDs, []domain.TagID{7, 8}) {
				t.Errorf("tagIDs = %v, want [7 8]", tagIDs)
			}
			livestreamTags := make([]*domain.LivestreamTag, len(livestreamIDs))
			for i, id := range livestreamIDs {
				livestreamTags[i] = &domain.LivestreamTag{LivestreamID: id}
			}
			return livestreamTags, err
		},
	}
}

func TestLivestreamUsecase_FindAllByTagName(t *testing.T) {
	f := testLivestreamFixture()
	var tagCalls int
	// ライブ配信 1 にはタグ 7, 8 の両方が付いているので、紐付けごとに 2 回現れる (移行前と同じ)
	livestreamTagRepo := newLivestreamTagRepositoryFindingByTagIDs(t, &tagCalls, []domain.LivestreamID{2, 1, 1}, nil)
	var livestreamCalls []domain.LivestreamID
	livestreamRepo := newLivestreamRepositoryWithModels(&livestreamCalls, testLivestreamModel1, testLivestreamModel2)
	u := newLivestreamUsecaseForTest(f, nil, newTagRepositoryFindingIDs(t, []domain.TagID{7, 8}, nil), livestreamRepo, livestreamTagRepo)

	got, err := u.FindAllByTagName(context.Background(), "ゲーム実況")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []*domain.LivestreamDetail{f.livestream(testLivestreamModel2), f.livestream(testLivestreamModel1), f.livestream(testLivestreamModel1)}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
	// 同じライブ配信は 1 回だけ引く
	if want := []domain.LivestreamID{2, 1}; !slices.Equal(livestreamCalls, want) {
		t.Errorf("livestream calls = %v, want %v", livestreamCalls, want)
	}
}

func TestLivestreamUsecase_FindAllByTagName_TagNotFound(t *testing.T) {
	var tagCalls int
	livestreamTagRepo := newLivestreamTagRepositoryFindingByTagIDs(t, &tagCalls, nil, nil)
	u := newLivestreamUsecaseForTest(testLivestreamFixture(), nil, newTagRepositoryFindingIDs(t, nil, nil), nil, livestreamTagRepo)

	got, err := u.FindAllByTagName(context.Background(), "ゲーム実況")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Errorf("got %#v, want empty non-nil slice", got)
	}
	// 空の IN () になるので紐付けの検索はしない
	if tagCalls != 0 {
		t.Errorf("livestream tag calls = %d, want 0", tagCalls)
	}
}

func TestLivestreamUsecase_FindAllByTagName_Errors(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name string
		// tagErr はタグの検索が返すエラー
		tagErr error
		// livestreamTagsErr は紐付けの検索が返すエラー
		livestreamTagsErr error
		// livestreamRepo はライブ配信を引く repository
		livestreamRepo *fakeLivestreamRepository
		wantErr        error
		wantMsg        string
	}{
		{name: "tag repository error", tagErr: boom, wantErr: boom, wantMsg: "failed to get tags: boom"},
		{name: "livestream tag repository error", livestreamTagsErr: boom, wantErr: boom, wantMsg: "failed to get livestreams: boom"},
		{
			name:           "livestream not found",
			livestreamRepo: newLivestreamRepositoryWithModels(nil),
			wantErr:        repository.ErrNotFound,
			wantMsg:        "failed to get livestreams: failed to get livestream 2: not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tagCalls int
			livestreamTagRepo := newLivestreamTagRepositoryFindingByTagIDs(t, &tagCalls, []domain.LivestreamID{2}, tt.livestreamTagsErr)
			u := newLivestreamUsecaseForTest(testLivestreamFixture(), nil, newTagRepositoryFindingIDs(t, []domain.TagID{7, 8}, tt.tagErr), tt.livestreamRepo, livestreamTagRepo)
			_, err := u.FindAllByTagName(context.Background(), "ゲーム実況")
			if !errors.Is(err, tt.wantErr) || err.Error() != tt.wantMsg {
				t.Errorf("err = %v, want %q", err, tt.wantMsg)
			}
		})
	}
}

// newLivestreamRepositoryForFindAll は一覧取得で呼ばれたメソッドを calls に記録し、models (失敗させる場合は err) を返す fakeLivestreamRepository を返す。
// limit 付きの場合はその値を gotLimit に取り出す。
func newLivestreamRepositoryForFindAll(calls *[]string, gotLimit *domain.Limit, models []*domain.Livestream, err error) *fakeLivestreamRepository {
	return &fakeLivestreamRepository{
		findAllOrderByIDDesc: func(context.Context, repository.Querier) ([]*domain.Livestream, error) {
			*calls = append(*calls, "FindAllOrderByIDDesc")
			return models, err
		},
		findAllOrderByIDDescLimited: func(_ context.Context, _ repository.Querier, limit domain.Limit) ([]*domain.Livestream, error) {
			*calls = append(*calls, "FindAllOrderByIDDescLimited")
			*gotLimit = limit
			return models, err
		},
	}
}

func TestLivestreamUsecase_FindAll(t *testing.T) {
	limit := domain.Limit(5)

	tests := []struct {
		name      string
		limit     *domain.Limit
		wantCalls []string
		wantLimit domain.Limit
	}{
		{name: "without limit", limit: nil, wantCalls: []string{"FindAllOrderByIDDesc"}},
		{name: "with limit", limit: &limit, wantCalls: []string{"FindAllOrderByIDDescLimited"}, wantLimit: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := testLivestreamFixture()
			var calls []string
			var gotLimit domain.Limit
			livestreamRepo := newLivestreamRepositoryForFindAll(&calls, &gotLimit, []*domain.Livestream{testLivestreamModel2, testLivestreamModel1}, nil)
			u := newLivestreamUsecaseForTest(f, nil, nil, livestreamRepo, nil)

			got, err := u.FindAll(context.Background(), tt.limit)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if want := []*domain.LivestreamDetail{f.livestream(testLivestreamModel2), f.livestream(testLivestreamModel1)}; !reflect.DeepEqual(got, want) {
				t.Errorf("got %+v, want %+v", got, want)
			}
			if !slices.Equal(calls, tt.wantCalls) {
				t.Errorf("calls = %v, want %v", calls, tt.wantCalls)
			}
			if gotLimit != tt.wantLimit {
				t.Errorf("limit = %d, want %d", gotLimit, tt.wantLimit)
			}
		})
	}
}

func TestLivestreamUsecase_FindAll_Error(t *testing.T) {
	boom := errors.New("boom")
	var calls []string
	var gotLimit domain.Limit
	livestreamRepo := newLivestreamRepositoryForFindAll(&calls, &gotLimit, nil, boom)
	u := newLivestreamUsecaseForTest(testLivestreamFixture(), nil, nil, livestreamRepo, nil)

	_, err := u.FindAll(context.Background(), nil)
	if want := "failed to get livestreams: boom"; !errors.Is(err, boom) || err.Error() != want {
		t.Errorf("err = %v, want %q", err, want)
	}
}
