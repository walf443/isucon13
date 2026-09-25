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
	repo := &fakeLivestreamRepository{livestream: want}
	u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, &fakeTagRepository{}, repo, &fakeReservationSlotRepository{}, &fakeLogger{})

	got, err := u.FindByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
	if repo.gotID != 1 {
		t.Errorf("id = %d, want 1", repo.gotID)
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
			u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, &fakeTagRepository{}, &fakeLivestreamRepository{err: tt.repoErr}, &fakeReservationSlotRepository{}, &fakeLogger{})
			_, err := u.FindByID(context.Background(), 1)
			tt.check(t, err)
		})
	}
}

func TestLivestreamUsecase_FindAllByUserID(t *testing.T) {
	want := []*model.Livestream{{ID: 1}, {ID: 2}}
	repo := &fakeLivestreamRepository{livestreams: want}
	u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, &fakeTagRepository{}, repo, &fakeReservationSlotRepository{}, &fakeLogger{})

	got, err := u.FindAllByUserID(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("got %+v, want %+v", got, want)
	}
	if repo.gotUserID != 42 {
		t.Errorf("userID = %d, want 42", repo.gotUserID)
	}
}

func TestLivestreamUsecase_FindAllByUserID_Error(t *testing.T) {
	boom := errors.New("boom")
	u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, &fakeTagRepository{}, &fakeLivestreamRepository{err: boom}, &fakeReservationSlotRepository{}, &fakeLogger{})

	_, err := u.FindAllByUserID(context.Background(), 42)
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
}

func TestLivestreamUsecase_FindAllByUsername(t *testing.T) {
	want := []*model.Livestream{{ID: 1}}
	userRepo := &fakeUserRepository{id: 42}
	livestreamRepo := &fakeLivestreamRepository{livestreams: want}
	u := NewLivestreamUsecase(&fakeTxManager{}, userRepo, &fakeTagRepository{}, livestreamRepo, &fakeReservationSlotRepository{}, &fakeLogger{})

	got, err := u.FindAllByUsername(context.Background(), "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0] != want[0] {
		t.Errorf("got %+v, want %+v", got, want)
	}
	if userRepo.gotName != "alice" {
		t.Errorf("name = %q, want %q", userRepo.gotName, "alice")
	}
	// ユーザ名から引いた ID で検索する
	if livestreamRepo.gotUserID != 42 {
		t.Errorf("userID = %d, want 42", livestreamRepo.gotUserID)
	}
}

func TestLivestreamUsecase_FindAllByUsername_Errors(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name           string
		userRepo       *fakeUserRepository
		livestreamRepo *fakeLivestreamRepository
		wantErr        error
	}{
		{
			name:           "user not found",
			userRepo:       &fakeUserRepository{err: repository.ErrNotFound},
			livestreamRepo: &fakeLivestreamRepository{},
			wantErr:        ErrUserNotFound,
		},
		{
			name:           "user repository error",
			userRepo:       &fakeUserRepository{err: boom},
			livestreamRepo: &fakeLivestreamRepository{},
			wantErr:        boom,
		},
		{
			name:           "livestream repository error",
			userRepo:       &fakeUserRepository{id: 42},
			livestreamRepo: &fakeLivestreamRepository{err: boom},
			wantErr:        boom,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewLivestreamUsecase(&fakeTxManager{}, tt.userRepo, &fakeTagRepository{}, tt.livestreamRepo, &fakeReservationSlotRepository{}, &fakeLogger{})
			_, err := u.FindAllByUsername(context.Background(), "alice")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestLivestreamUsecase_FindAllByTagName(t *testing.T) {
	want := []*model.Livestream{{ID: 2}, {ID: 1}}
	tagRepo := &fakeTagRepository{ids: []model.TagID{7}}
	livestreamRepo := &fakeLivestreamRepository{livestreams: want}
	u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, tagRepo, livestreamRepo, &fakeReservationSlotRepository{}, &fakeLogger{})

	got, err := u.FindAllByTagName(context.Background(), "ゲーム実況")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !slices.Equal(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
	if tagRepo.gotName != "ゲーム実況" {
		t.Errorf("tag name = %q", tagRepo.gotName)
	}
	if !slices.Equal(livestreamRepo.gotTagIDs, []model.TagID{7}) {
		t.Errorf("tagIDs = %v, want [7]", livestreamRepo.gotTagIDs)
	}
}

func TestLivestreamUsecase_FindAllByTagName_TagNotFound(t *testing.T) {
	livestreamRepo := &fakeLivestreamRepository{}
	u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, &fakeTagRepository{ids: nil}, livestreamRepo, &fakeReservationSlotRepository{}, &fakeLogger{})

	got, err := u.FindAllByTagName(context.Background(), "nothing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Errorf("got %#v, want empty non-nil slice", got)
	}
	// 空の IN () になるので livestream の検索はしない
	if len(livestreamRepo.calls) != 0 {
		t.Errorf("calls = %v, want none", livestreamRepo.calls)
	}
}

func TestLivestreamUsecase_FindAllByTagName_Errors(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name           string
		tagRepo        *fakeTagRepository
		livestreamRepo *fakeLivestreamRepository
	}{
		{name: "tag repository error", tagRepo: &fakeTagRepository{err: boom}, livestreamRepo: &fakeLivestreamRepository{}},
		{name: "livestream repository error", tagRepo: &fakeTagRepository{ids: []model.TagID{7}}, livestreamRepo: &fakeLivestreamRepository{err: boom}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, tt.tagRepo, tt.livestreamRepo, &fakeReservationSlotRepository{}, &fakeLogger{})
			_, err := u.FindAllByTagName(context.Background(), "ゲーム実況")
			if !errors.Is(err, boom) {
				t.Fatalf("err = %v, want %v", err, boom)
			}
		})
	}
}

func TestLivestreamUsecase_FindAll(t *testing.T) {
	limit := int64(5)

	tests := []struct {
		name      string
		limit     *int64
		wantCalls []string
		wantLimit int64
	}{
		{name: "without limit", limit: nil, wantCalls: []string{"FindAllWithDetails"}},
		{name: "with limit", limit: &limit, wantCalls: []string{"FindAllWithDetailsLimited"}, wantLimit: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := []*model.Livestream{{ID: 1}}
			livestreamRepo := &fakeLivestreamRepository{livestreams: want}
			u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, &fakeTagRepository{}, livestreamRepo, &fakeReservationSlotRepository{}, &fakeLogger{})

			got, err := u.FindAll(context.Background(), tt.limit)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !slices.Equal(got, want) {
				t.Errorf("got %+v, want %+v", got, want)
			}
			if !slices.Equal(livestreamRepo.calls, tt.wantCalls) {
				t.Errorf("calls = %v, want %v", livestreamRepo.calls, tt.wantCalls)
			}
			if livestreamRepo.gotLimit != tt.wantLimit {
				t.Errorf("limit = %d, want %d", livestreamRepo.gotLimit, tt.wantLimit)
			}
		})
	}
}

func TestLivestreamUsecase_FindAll_Error(t *testing.T) {
	boom := errors.New("boom")
	u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, &fakeTagRepository{}, &fakeLivestreamRepository{err: boom}, &fakeReservationSlotRepository{}, &fakeLogger{})

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

func TestLivestreamUsecase_Reserve(t *testing.T) {
	want := &model.Livestream{ID: 100, Title: "stream"}
	livestreamRepo := &fakeLivestreamRepository{createID: 100, livestream: want}
	slotRepo := &fakeReservationSlotRepository{
		slots: []*model.ReservationSlotModel{
			{Slot: 5, StartAt: 1700874000, EndAt: 1700877600},
			{Slot: 3, StartAt: 1700877600, EndAt: 1700881200},
		},
		counts: map[int64]int64{1700874000: 5, 1700877600: 3},
	}
	logger := &fakeLogger{}
	u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, &fakeTagRepository{}, livestreamRepo, slotRepo, logger)

	got, err := u.Reserve(context.Background(), 1, testReserveInput)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
	if want := []string{"FindAllByRangeForUpdate", "FindSlotByStartAtAndEndAt", "FindSlotByStartAtAndEndAt", "DecrementSlotsByRange"}; !slices.Equal(slotRepo.calls, want) {
		t.Errorf("slot calls = %v, want %v", slotRepo.calls, want)
	}
	if slotRepo.gotStartAt != 1700874000 || slotRepo.gotEndAt != 1700881200 {
		t.Errorf("slot range = %d ~ %d", slotRepo.gotStartAt, slotRepo.gotEndAt)
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
	if *livestreamRepo.gotCreated != wantCreated {
		t.Errorf("created = %+v, want %+v", *livestreamRepo.gotCreated, wantCreated)
	}
	if want := []string{"Create", "AddTag", "AddTag", "FindWithDetailsByID"}; !slices.Equal(livestreamRepo.calls, want) {
		t.Errorf("livestream calls = %v, want %v", livestreamRepo.calls, want)
	}
	if want := []model.TagID{3, 5}; !slices.Equal(livestreamRepo.addedTagIDs, want) {
		t.Errorf("tag ids = %v, want %v", livestreamRepo.addedTagIDs, want)
	}
	if livestreamRepo.gotID != 100 {
		t.Errorf("re-read id = %d, want 100", livestreamRepo.gotID)
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
			slotRepo := &fakeReservationSlotRepository{}
			u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, &fakeTagRepository{}, &fakeLivestreamRepository{}, slotRepo, &fakeLogger{})

			input := testReserveInput
			input.StartAt = tt.startAt
			input.EndAt = tt.endAt
			_, err := u.Reserve(context.Background(), 1, input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil && len(slotRepo.calls) != 0 {
				t.Errorf("slot calls = %v, want none", slotRepo.calls)
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
		name           string
		slotRepo       *fakeReservationSlotRepository
		livestreamRepo *fakeLivestreamRepository
		wantErr        error
		// wantUnavailable は *ReservationSlotUnavailableError を期待するかどうか
		wantUnavailable     bool
		wantLivestreamCalls []string
		wantLogLines        []string
		wantWarnLines       []string
	}{
		{
			name:           "slot list fails",
			slotRepo:       &fakeReservationSlotRepository{findAllErr: boom},
			livestreamRepo: &fakeLivestreamRepository{},
			wantErr:        boom,
			wantWarnLines:  []string{"予約枠一覧取得でエラー発生: boom"},
		},
		{
			name:           "slot count fails",
			slotRepo:       &fakeReservationSlotRepository{slots: slots, findSlotErr: boom},
			livestreamRepo: &fakeLivestreamRepository{},
			wantErr:        boom,
		},
		{
			// 残数の判定は FOR UPDATE で取得した値ではなく、取り直した値で行う (移行前と同じ)
			name:            "slot is full",
			slotRepo:        &fakeReservationSlotRepository{slots: slots, counts: map[int64]int64{1700874000: 5, 1700877600: 0}},
			livestreamRepo:  &fakeLivestreamRepository{},
			wantUnavailable: true,
			wantLogLines:    []string{"1700874000 ~ 1700877600予約枠の残数 = 5\n", "1700877600 ~ 1700881200予約枠の残数 = 3\n"},
		},
		{
			name:           "decrement fails",
			slotRepo:       &fakeReservationSlotRepository{slots: slots, counts: available, decrementErr: boom},
			livestreamRepo: &fakeLivestreamRepository{},
			wantErr:        boom,
			wantLogLines:   []string{"1700874000 ~ 1700877600予約枠の残数 = 5\n", "1700877600 ~ 1700881200予約枠の残数 = 3\n"},
		},
		{
			name:                "create fails",
			slotRepo:            &fakeReservationSlotRepository{slots: slots, counts: available},
			livestreamRepo:      &fakeLivestreamRepository{createErr: boom},
			wantErr:             boom,
			wantLivestreamCalls: []string{"Create"},
			wantLogLines:        []string{"1700874000 ~ 1700877600予約枠の残数 = 5\n", "1700877600 ~ 1700881200予約枠の残数 = 3\n"},
		},
		{
			name:                "add tag fails",
			slotRepo:            &fakeReservationSlotRepository{slots: slots, counts: available},
			livestreamRepo:      &fakeLivestreamRepository{createID: 100, addTagErr: boom},
			wantErr:             boom,
			wantLivestreamCalls: []string{"Create", "AddTag"},
			wantLogLines:        []string{"1700874000 ~ 1700877600予約枠の残数 = 5\n", "1700877600 ~ 1700881200予約枠の残数 = 3\n"},
		},
		{
			name:                "fill fails",
			slotRepo:            &fakeReservationSlotRepository{slots: slots, counts: available},
			livestreamRepo:      &fakeLivestreamRepository{createID: 100, err: boom},
			wantErr:             boom,
			wantLivestreamCalls: []string{"Create", "AddTag", "AddTag", "FindWithDetailsByID"},
			wantLogLines:        []string{"1700874000 ~ 1700877600予約枠の残数 = 5\n", "1700877600 ~ 1700881200予約枠の残数 = 3\n"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &fakeLogger{}
			u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, &fakeTagRepository{}, tt.livestreamRepo, tt.slotRepo, logger)
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
			if !slices.Equal(tt.livestreamRepo.calls, tt.wantLivestreamCalls) {
				t.Errorf("livestream calls = %v, want %v", tt.livestreamRepo.calls, tt.wantLivestreamCalls)
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
		name     string
		slotRepo *fakeReservationSlotRepository
		want     string
	}{
		{name: "slot list", slotRepo: &fakeReservationSlotRepository{findAllErr: boom}, want: "failed to get reservation_slots: boom"},
		{name: "slot count", slotRepo: &fakeReservationSlotRepository{slots: []*model.ReservationSlotModel{{StartAt: 1700874000, EndAt: 1700877600}}, findSlotErr: boom}, want: "failed to get reservation_slots: boom"},
		{name: "decrement", slotRepo: &fakeReservationSlotRepository{decrementErr: boom}, want: "failed to update reservation_slot: boom"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewLivestreamUsecase(&fakeTxManager{}, &fakeUserRepository{}, &fakeTagRepository{}, &fakeLivestreamRepository{}, tt.slotRepo, &fakeLogger{})
			_, err := u.Reserve(context.Background(), 1, testReserveInput)
			if err == nil || err.Error() != tt.want {
				t.Errorf("err = %v, want %q", err, tt.want)
			}
		})
	}
}
