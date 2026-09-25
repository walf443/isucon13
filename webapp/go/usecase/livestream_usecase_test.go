package usecase

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

func TestLivestreamUsecase_FindByID(t *testing.T) {
	want := &model.Livestream{ID: 1, Title: "stream"}
	repo := &fakeLivestreamRepository{
		findWithDetailsByID: func(_ context.Context, _ repository.Querier, id model.LivestreamID) (*model.Livestream, error) {
			if id != 1 {
				t.Errorf("id = %d, want 1", id)
			}
			return want, nil
		},
	}
	u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, &fakeTagRepository{}, repo, &fakeReservationSlotRepository{}, &fakeLogger{})

	got, err := u.FindByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestLivestreamUsecase_FindByID_Errors(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name    string
		repoErr error
		check   func(t *testing.T, err error)
	}{
		{
			name:    "livestream not found",
			repoErr: repository.ErrNotFound,
			check: func(t *testing.T, err error) {
				if !errors.Is(err, ErrLivestreamNotFound) {
					t.Errorf("err = %v, want ErrLivestreamNotFound", err)
				}
			},
		},
		{
			name:    "unexpected error",
			repoErr: boom,
			check: func(t *testing.T, err error) {
				if !errors.Is(err, boom) || errors.Is(err, ErrLivestreamNotFound) {
					t.Errorf("err = %v, want wrapped %v", err, boom)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeLivestreamRepository{
				findWithDetailsByID: func(context.Context, repository.Querier, model.LivestreamID) (*model.Livestream, error) {
					return nil, tt.repoErr
				},
			}
			u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, &fakeTagRepository{}, repo, &fakeReservationSlotRepository{}, &fakeLogger{})
			_, err := u.FindByID(context.Background(), 1)
			tt.check(t, err)
		})
	}
}

// newLivestreamRepositoryFindingAllByUserID は配信者 userID のライブ配信として livestreams (失敗させる場合は err) を返す fakeLivestreamRepository を返す。
func newLivestreamRepositoryFindingAllByUserID(t *testing.T, userID model.UserID, livestreams []*model.Livestream, err error) *fakeLivestreamRepository {
	return &fakeLivestreamRepository{
		findAllWithDetailsByUserID: func(_ context.Context, _ repository.Querier, gotUserID model.UserID) ([]*model.Livestream, error) {
			if gotUserID != userID {
				t.Errorf("userID = %d, want %d", gotUserID, userID)
			}
			return livestreams, err
		},
	}
}

func TestLivestreamUsecase_FindAllByUserID(t *testing.T) {
	want := []*model.Livestream{{ID: 1}, {ID: 2}}
	repo := newLivestreamRepositoryFindingAllByUserID(t, 42, want, nil)
	u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, &fakeTagRepository{}, repo, &fakeReservationSlotRepository{}, &fakeLogger{})

	got, err := u.FindAllByUserID(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestLivestreamUsecase_FindAllByUserID_Error(t *testing.T) {
	boom := errors.New("boom")
	repo := newLivestreamRepositoryFindingAllByUserID(t, 42, nil, boom)
	u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, &fakeTagRepository{}, repo, &fakeReservationSlotRepository{}, &fakeLogger{})

	_, err := u.FindAllByUserID(context.Background(), 42)
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
}

func TestLivestreamUsecase_FindAllByUsername(t *testing.T) {
	want := []*model.Livestream{{ID: 1}}
	// ユーザ名から引いた ID で検索する
	livestreamRepo := newLivestreamRepositoryFindingAllByUserID(t, 42, want, nil)
	u := NewLivestreamUsecase(&fakeTxManager{}, newUserRepositoryFindingID(t, "alice", 42, nil), &fakeTagRepository{}, livestreamRepo, &fakeReservationSlotRepository{}, &fakeLogger{})

	got, err := u.FindAllByUsername(context.Background(), "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0] != want[0] {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestLivestreamUsecase_FindAllByUsername_Errors(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name string
		// userID, userErr はユーザ名から ID を引いた結果
		userID  model.UserID
		userErr error
		// livestreamsErr はライブ配信の検索が返すエラー
		livestreamsErr error
		wantErr        error
	}{
		{
			name:    "user not found",
			userErr: repository.ErrNotFound,
			wantErr: ErrUserNotFound,
		},
		{
			name:    "user repository error",
			userErr: boom,
			wantErr: boom,
		},
		{
			name:           "livestream repository error",
			userID:         42,
			livestreamsErr: boom,
			wantErr:        boom,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			livestreamRepo := newLivestreamRepositoryFindingAllByUserID(t, 42, nil, tt.livestreamsErr)
			u := NewLivestreamUsecase(&fakeTxManager{}, newUserRepositoryFindingID(t, "alice", tt.userID, tt.userErr), &fakeTagRepository{}, livestreamRepo, &fakeReservationSlotRepository{}, &fakeLogger{})
			_, err := u.FindAllByUsername(context.Background(), "alice")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

// newLivestreamRepositoryFindingAllByTagIDs はタグ 7 のライブ配信として livestreams (失敗させる場合は err) を返す fakeLivestreamRepository を返す。
// 呼ばれた回数を calls に数える。
func newLivestreamRepositoryFindingAllByTagIDs(t *testing.T, calls *int, livestreams []*model.Livestream, err error) *fakeLivestreamRepository {
	return &fakeLivestreamRepository{
		findAllWithDetailsByTagIDs: func(_ context.Context, _ repository.Querier, tagIDs []model.TagID) ([]*model.Livestream, error) {
			*calls++
			if !slices.Equal(tagIDs, []model.TagID{7}) {
				t.Errorf("tagIDs = %v, want [7]", tagIDs)
			}
			return livestreams, err
		},
	}
}

func TestLivestreamUsecase_FindAllByTagName(t *testing.T) {
	want := []*model.Livestream{{ID: 2}, {ID: 1}}
	tagRepo := &fakeTagRepository{
		findIDsByName: func(_ context.Context, _ repository.Querier, name string) ([]model.TagID, error) {
			if name != "ゲーム実況" {
				t.Errorf("tag name = %q", name)
			}
			return []model.TagID{7}, nil
		},
	}
	var livestreamCalls int
	livestreamRepo := newLivestreamRepositoryFindingAllByTagIDs(t, &livestreamCalls, want, nil)
	u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, tagRepo, livestreamRepo, &fakeReservationSlotRepository{}, &fakeLogger{})

	got, err := u.FindAllByTagName(context.Background(), "ゲーム実況")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !slices.Equal(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestLivestreamUsecase_FindAllByTagName_TagNotFound(t *testing.T) {
	var livestreamCalls int
	livestreamRepo := newLivestreamRepositoryFindingAllByTagIDs(t, &livestreamCalls, nil, nil)
	tagRepo := &fakeTagRepository{
		findIDsByName: func(context.Context, repository.Querier, string) ([]model.TagID, error) { return nil, nil },
	}
	u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, tagRepo, livestreamRepo, &fakeReservationSlotRepository{}, &fakeLogger{})

	got, err := u.FindAllByTagName(context.Background(), "nothing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Errorf("got %#v, want empty non-nil slice", got)
	}
	// 空の IN () になるので livestream の検索はしない
	if livestreamCalls != 0 {
		t.Errorf("livestream calls = %d, want 0", livestreamCalls)
	}
}

func TestLivestreamUsecase_FindAllByTagName_Errors(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name string
		// tagIDs, tagErr はタグの検索が返す値
		tagIDs []model.TagID
		tagErr error
		// livestreamsErr はライブ配信の検索が返すエラー
		livestreamsErr error
	}{
		{name: "tag repository error", tagErr: boom},
		{name: "livestream repository error", tagIDs: []model.TagID{7}, livestreamsErr: boom},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tagRepo := &fakeTagRepository{
				findIDsByName: func(context.Context, repository.Querier, string) ([]model.TagID, error) { return tt.tagIDs, tt.tagErr },
			}
			var livestreamCalls int
			livestreamRepo := newLivestreamRepositoryFindingAllByTagIDs(t, &livestreamCalls, nil, tt.livestreamsErr)
			u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, tagRepo, livestreamRepo, &fakeReservationSlotRepository{}, &fakeLogger{})
			_, err := u.FindAllByTagName(context.Background(), "ゲーム実況")
			if !errors.Is(err, boom) {
				t.Fatalf("err = %v, want %v", err, boom)
			}
		})
	}
}

// newLivestreamRepositoryForFindAll は一覧取得で呼ばれたメソッドを calls に記録し、livestreams (失敗させる場合は err) を返す fakeLivestreamRepository を返す。
// limit 付きの場合はその値を gotLimit に取り出す。
func newLivestreamRepositoryForFindAll(calls *[]string, gotLimit *model.Limit, livestreams []*model.Livestream, err error) *fakeLivestreamRepository {
	return &fakeLivestreamRepository{
		findAllWithDetails: func(context.Context, repository.Querier) ([]*model.Livestream, error) {
			*calls = append(*calls, "FindAllWithDetails")
			return livestreams, err
		},
		findAllWithDetailsLimited: func(_ context.Context, _ repository.Querier, limit model.Limit) ([]*model.Livestream, error) {
			*calls = append(*calls, "FindAllWithDetailsLimited")
			*gotLimit = limit
			return livestreams, err
		},
	}
}

func TestLivestreamUsecase_FindAll(t *testing.T) {
	limit := model.Limit(5)

	tests := []struct {
		name      string
		limit     *model.Limit
		wantCalls []string
		wantLimit model.Limit
	}{
		{name: "without limit", limit: nil, wantCalls: []string{"FindAllWithDetails"}},
		{name: "with limit", limit: &limit, wantCalls: []string{"FindAllWithDetailsLimited"}, wantLimit: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := []*model.Livestream{{ID: 1}}
			var calls []string
			var gotLimit model.Limit
			livestreamRepo := newLivestreamRepositoryForFindAll(&calls, &gotLimit, want, nil)
			u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, &fakeTagRepository{}, livestreamRepo, &fakeReservationSlotRepository{}, &fakeLogger{})

			got, err := u.FindAll(context.Background(), tt.limit)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !slices.Equal(got, want) {
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
	var gotLimit model.Limit
	livestreamRepo := newLivestreamRepositoryForFindAll(&calls, &gotLimit, nil, boom)
	u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, &fakeTagRepository{}, livestreamRepo, &fakeReservationSlotRepository{}, &fakeLogger{})

	_, err := u.FindAll(context.Background(), nil)
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
}

// 予約可能期間は 1700874000 (2023/11/25 10:00 JST) 〜 1732496400 (2024/11/25 10:00 JST)
var testReserveInput = ReserveLivestreamInput{
	TagIDs:       []model.TagID{3, 5},
	Title:        "stream",
	Description:  "desc",
	PlaylistUrl:  "https://example.com/playlist.m3u8",
	ThumbnailUrl: "https://example.com/thumbnail.png",
	StartAt:      1700874000,
	EndAt:        1700881200,
}

// reservationSlotResults は予約の処理で予約枠の repository が返す値。
type reservationSlotResults struct {
	slots []*model.ReservationSlotModel
	// counts は FindSlotByStartAtAndEndAt が返す残数 (キーは開始時刻)
	counts       map[int64]int64
	findAllErr   error
	findSlotErr  error
	decrementErr error
}

// newReservationSlotRepositoryForReserve は results を返し、呼ばれたメソッドを順に calls へ記録する fakeReservationSlotRepository を返す。
// 予約枠の検索・残数の減算は予約区間 startAt 〜 endAt で呼ばれることを確認する。
func newReservationSlotRepositoryForReserve(t *testing.T, calls *[]string, results reservationSlotResults, startAt, endAt int64) *fakeReservationSlotRepository {
	checkRange := func(method string, gotStartAt, gotEndAt int64) {
		if gotStartAt != startAt || gotEndAt != endAt {
			t.Errorf("%s range = %d ~ %d, want %d ~ %d", method, gotStartAt, gotEndAt, startAt, endAt)
		}
	}
	return &fakeReservationSlotRepository{
		findAllByRangeForUpdate: func(_ context.Context, _ repository.Querier, gotStartAt int64, gotEndAt int64) ([]*model.ReservationSlotModel, error) {
			*calls = append(*calls, "FindAllByRangeForUpdate")
			checkRange("FindAllByRangeForUpdate", gotStartAt, gotEndAt)
			return results.slots, results.findAllErr
		},
		findSlotByStartAtAndEndAt: func(_ context.Context, _ repository.Querier, slotStartAt int64, _ int64) (int64, error) {
			*calls = append(*calls, "FindSlotByStartAtAndEndAt")
			return results.counts[slotStartAt], results.findSlotErr
		},
		decrementSlotsByRange: func(_ context.Context, _ repository.Querier, gotStartAt int64, gotEndAt int64) error {
			*calls = append(*calls, "DecrementSlotsByRange")
			checkRange("DecrementSlotsByRange", gotStartAt, gotEndAt)
			return results.decrementErr
		},
	}
}

// livestreamReserveResults は予約の処理でライブ配信の repository が返す値。
type livestreamReserveResults struct {
	// livestream は ID 100 で読み直したライブ配信
	livestream *model.Livestream
	createErr  error
	addTagErr  error
	fillErr    error
}

// newLivestreamRepositoryForReserve は results を返し、呼ばれたメソッドを順に calls へ記録する fakeLivestreamRepository を返す。
// 登録したライブ配信は created に、付けたタグの ID は addedTagIDs に取り出す。登録したライブ配信の ID は 100 とする。
func newLivestreamRepositoryForReserve(t *testing.T, calls *[]string, created **model.LivestreamModel, addedTagIDs *[]model.TagID, results livestreamReserveResults) *fakeLivestreamRepository {
	return &fakeLivestreamRepository{
		create: func(_ context.Context, _ repository.Querier, livestream *model.LivestreamModel) (model.LivestreamID, error) {
			*calls = append(*calls, "Create")
			*created = livestream
			return 100, results.createErr
		},
		addTag: func(_ context.Context, _ repository.Querier, livestreamID model.LivestreamID, tagID model.TagID) error {
			*calls = append(*calls, "AddTag")
			*addedTagIDs = append(*addedTagIDs, tagID)
			if livestreamID != 100 {
				t.Errorf("tag added to livestream %d, want 100", livestreamID)
			}
			return results.addTagErr
		},
		findWithDetailsByID: func(_ context.Context, _ repository.Querier, id model.LivestreamID) (*model.Livestream, error) {
			*calls = append(*calls, "FindWithDetailsByID")
			if id != 100 {
				t.Errorf("re-read id = %d, want 100", id)
			}
			return results.livestream, results.fillErr
		},
	}
}

func TestLivestreamUsecase_Reserve(t *testing.T) {
	want := &model.Livestream{ID: 100, Title: "stream"}
	var livestreamCalls []string
	var created *model.LivestreamModel
	var addedTagIDs []model.TagID
	livestreamRepo := newLivestreamRepositoryForReserve(t, &livestreamCalls, &created, &addedTagIDs, livestreamReserveResults{livestream: want})
	var slotCalls []string
	slotRepo := newReservationSlotRepositoryForReserve(t, &slotCalls, reservationSlotResults{
		slots: []*model.ReservationSlotModel{
			{Slot: 5, StartAt: 1700874000, EndAt: 1700877600},
			{Slot: 3, StartAt: 1700877600, EndAt: 1700881200},
		},
		counts: map[int64]int64{1700874000: 5, 1700877600: 3},
	}, testReserveInput.StartAt, testReserveInput.EndAt)
	logger := &fakeLogger{}
	u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, &fakeTagRepository{}, livestreamRepo, slotRepo, logger)

	got, err := u.Reserve(context.Background(), 1, testReserveInput)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
	if want := []string{"FindAllByRangeForUpdate", "FindSlotByStartAtAndEndAt", "FindSlotByStartAtAndEndAt", "DecrementSlotsByRange"}; !slices.Equal(slotCalls, want) {
		t.Errorf("slot calls = %v, want %v", slotCalls, want)
	}
	// 予約枠ごとに残数をログに出す (移行前と同じ形式で、末尾の改行も含む)
	if want := []string{"1700874000 ~ 1700877600予約枠の残数 = 5\n", "1700877600 ~ 1700881200予約枠の残数 = 3\n"}; !slices.Equal(logger.lines, want) {
		t.Errorf("log lines = %q, want %q", logger.lines, want)
	}
	wantCreated := model.LivestreamModel{
		UserID:       1,
		Title:        "stream",
		Description:  "desc",
		PlaylistUrl:  "https://example.com/playlist.m3u8",
		ThumbnailUrl: "https://example.com/thumbnail.png",
		StartAt:      1700874000,
		EndAt:        1700881200,
	}
	if created == nil || *created != wantCreated {
		t.Errorf("created = %+v, want %+v", created, wantCreated)
	}
	if want := []string{"Create", "AddTag", "AddTag", "FindWithDetailsByID"}; !slices.Equal(livestreamCalls, want) {
		t.Errorf("livestream calls = %v, want %v", livestreamCalls, want)
	}
	if want := []model.TagID{3, 5}; !slices.Equal(addedTagIDs, want) {
		t.Errorf("tag ids = %v, want %v", addedTagIDs, want)
	}
}

func TestLivestreamUsecase_Reserve_TimeRange(t *testing.T) {
	tests := []struct {
		name    string
		startAt int64
		endAt   int64
		wantErr error
	}{
		{name: "ends at the term start", startAt: 1700870400, endAt: 1700874000, wantErr: ErrBadReservationTimeRange},
		{name: "ends before the term start", startAt: 1700866800, endAt: 1700870400, wantErr: ErrBadReservationTimeRange},
		{name: "starts at the term end", startAt: 1732496400, endAt: 1732500000, wantErr: ErrBadReservationTimeRange},
		{name: "starts after the term end", startAt: 1732500000, endAt: 1732503600, wantErr: ErrBadReservationTimeRange},
		// 期間に一部でも掛かっていれば予約できる (移行前と同じ)
		{name: "overlaps the term start", startAt: 1700870400, endAt: 1700877600, wantErr: nil},
		{name: "overlaps the term end", startAt: 1732492800, endAt: 1732500000, wantErr: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var slotCalls []string
			slotRepo := newReservationSlotRepositoryForReserve(t, &slotCalls, reservationSlotResults{}, tt.startAt, tt.endAt)
			var livestreamCalls []string
			var created *model.LivestreamModel
			var addedTagIDs []model.TagID
			livestreamRepo := newLivestreamRepositoryForReserve(t, &livestreamCalls, &created, &addedTagIDs, livestreamReserveResults{})
			u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, &fakeTagRepository{}, livestreamRepo, slotRepo, &fakeLogger{})

			input := testReserveInput
			input.StartAt = tt.startAt
			input.EndAt = tt.endAt
			_, err := u.Reserve(context.Background(), 1, input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil && len(slotCalls) != 0 {
				t.Errorf("slot calls = %v, want none", slotCalls)
			}
		})
	}
}

func TestLivestreamUsecase_Reserve_Errors(t *testing.T) {
	boom := errors.New("boom")
	slots := []*model.ReservationSlotModel{
		{Slot: 5, StartAt: 1700874000, EndAt: 1700877600},
		{Slot: 3, StartAt: 1700877600, EndAt: 1700881200},
	}
	available := map[int64]int64{1700874000: 5, 1700877600: 3}

	tests := []struct {
		name              string
		slotResults       reservationSlotResults
		livestreamResults livestreamReserveResults
		wantErr           error
		// wantUnavailable は *ReservationSlotUnavailableError を期待するかどうか
		wantUnavailable     bool
		wantLivestreamCalls []string
		wantLogLines        []string
		wantWarnLines       []string
	}{
		{
			name:          "slot list fails",
			slotResults:   reservationSlotResults{findAllErr: boom},
			wantErr:       boom,
			wantWarnLines: []string{"予約枠一覧取得でエラー発生: boom"},
		},
		{
			name:        "slot count fails",
			slotResults: reservationSlotResults{slots: slots, findSlotErr: boom},
			wantErr:     boom,
		},
		{
			// 残数の判定は FOR UPDATE で取得した値ではなく、取り直した値で行う (移行前と同じ)
			name:            "slot is full",
			slotResults:     reservationSlotResults{slots: slots, counts: map[int64]int64{1700874000: 5, 1700877600: 0}},
			wantUnavailable: true,
			wantLogLines:    []string{"1700874000 ~ 1700877600予約枠の残数 = 5\n", "1700877600 ~ 1700881200予約枠の残数 = 3\n"},
		},
		{
			name:         "decrement fails",
			slotResults:  reservationSlotResults{slots: slots, counts: available, decrementErr: boom},
			wantErr:      boom,
			wantLogLines: []string{"1700874000 ~ 1700877600予約枠の残数 = 5\n", "1700877600 ~ 1700881200予約枠の残数 = 3\n"},
		},
		{
			name:                "create fails",
			slotResults:         reservationSlotResults{slots: slots, counts: available},
			livestreamResults:   livestreamReserveResults{createErr: boom},
			wantErr:             boom,
			wantLivestreamCalls: []string{"Create"},
			wantLogLines:        []string{"1700874000 ~ 1700877600予約枠の残数 = 5\n", "1700877600 ~ 1700881200予約枠の残数 = 3\n"},
		},
		{
			name:                "add tag fails",
			slotResults:         reservationSlotResults{slots: slots, counts: available},
			livestreamResults:   livestreamReserveResults{addTagErr: boom},
			wantErr:             boom,
			wantLivestreamCalls: []string{"Create", "AddTag"},
			wantLogLines:        []string{"1700874000 ~ 1700877600予約枠の残数 = 5\n", "1700877600 ~ 1700881200予約枠の残数 = 3\n"},
		},
		{
			name:                "fill fails",
			slotResults:         reservationSlotResults{slots: slots, counts: available},
			livestreamResults:   livestreamReserveResults{fillErr: boom},
			wantErr:             boom,
			wantLivestreamCalls: []string{"Create", "AddTag", "AddTag", "FindWithDetailsByID"},
			wantLogLines:        []string{"1700874000 ~ 1700877600予約枠の残数 = 5\n", "1700877600 ~ 1700881200予約枠の残数 = 3\n"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &fakeLogger{}
			var slotCalls []string
			slotRepo := newReservationSlotRepositoryForReserve(t, &slotCalls, tt.slotResults, testReserveInput.StartAt, testReserveInput.EndAt)
			var livestreamCalls []string
			var created *model.LivestreamModel
			var addedTagIDs []model.TagID
			livestreamRepo := newLivestreamRepositoryForReserve(t, &livestreamCalls, &created, &addedTagIDs, tt.livestreamResults)
			u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, &fakeTagRepository{}, livestreamRepo, slotRepo, logger)
			_, err := u.Reserve(context.Background(), 1, testReserveInput)
			if tt.wantUnavailable {
				unavailable, ok := errors.AsType[*ReservationSlotUnavailableError](err)
				if !ok {
					t.Fatalf("err = %v, want *ReservationSlotUnavailableError", err)
				}
				if unavailable.StartAt != testReserveInput.StartAt || unavailable.EndAt != testReserveInput.EndAt {
					t.Errorf("unavailable = %+v", unavailable)
				}
				// メッセージには予約可能期間とリクエストの予約区間を含める (移行前と同じ)
				if want := "予約期間 1700874000 ~ 1732496400に対して、予約区間 1700874000 ~ 1700881200が予約できません"; err.Error() != want {
					t.Errorf("err = %q, want %q", err.Error(), want)
				}
			} else if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			// 予約枠を確保できなかった場合はライブ配信を登録しない
			if !slices.Equal(livestreamCalls, tt.wantLivestreamCalls) {
				t.Errorf("livestream calls = %v, want %v", livestreamCalls, tt.wantLivestreamCalls)
			}
			if !slices.Equal(logger.lines, tt.wantLogLines) {
				t.Errorf("log lines = %q, want %q", logger.lines, tt.wantLogLines)
			}
			if !slices.Equal(logger.warnLines, tt.wantWarnLines) {
				t.Errorf("warn lines = %q, want %q", logger.warnLines, tt.wantWarnLines)
			}
		})
	}
}

func TestLivestreamUsecase_Reserve_ErrorMessages(t *testing.T) {
	boom := errors.New("boom")
	tests := []struct {
		name        string
		slotResults reservationSlotResults
		want        string
	}{
		{name: "slot list", slotResults: reservationSlotResults{findAllErr: boom}, want: "failed to get reservation_slots: boom"},
		{name: "slot count", slotResults: reservationSlotResults{slots: []*model.ReservationSlotModel{{StartAt: 1700874000, EndAt: 1700877600}}, findSlotErr: boom}, want: "failed to get reservation_slots: boom"},
		{name: "decrement", slotResults: reservationSlotResults{decrementErr: boom}, want: "failed to update reservation_slot: boom"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var slotCalls []string
			slotRepo := newReservationSlotRepositoryForReserve(t, &slotCalls, tt.slotResults, testReserveInput.StartAt, testReserveInput.EndAt)
			u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, &fakeTagRepository{}, &fakeLivestreamRepository{}, slotRepo, &fakeLogger{})
			_, err := u.Reserve(context.Background(), 1, testReserveInput)
			if err == nil || err.Error() != tt.want {
				t.Errorf("err = %v, want %q", err, tt.want)
			}
		})
	}
}
