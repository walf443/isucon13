package usecase

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

// testReservePeriod は testReserveInput の予約区間。
var testReservePeriod = model.ReservationPeriod{StartAt: 1700874000, EndAt: 1700881200}

// 予約可能期間は 1700874000 (2023/11/25 10:00 JST) 〜 1732496400 (2024/11/25 10:00 JST)
var testReserveInput = ReserveLivestreamInput{
	TagIDs:       []model.TagID{3, 5},
	Title:        "stream",
	Description:  "desc",
	PlaylistUrl:  "https://example.com/playlist.m3u8",
	ThumbnailUrl: "https://example.com/thumbnail.png",
	StartAt:      testReservePeriod.StartAt,
	EndAt:        testReservePeriod.EndAt,
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
// 予約枠の検索・残数の減算は予約区間 period で呼ばれることを確認する。
func newReservationSlotRepositoryForReserve(t *testing.T, calls *[]string, results reservationSlotResults, period model.ReservationPeriod) *fakeReservationSlotRepository {
	checkPeriod := func(method string, got model.ReservationPeriod) {
		if got != period {
			t.Errorf("%s period = %+v, want %+v", method, got, period)
		}
	}
	return &fakeReservationSlotRepository{
		findAllByRangeForUpdate: func(_ context.Context, _ repository.Querier, got model.ReservationPeriod) ([]*model.ReservationSlotModel, error) {
			*calls = append(*calls, "FindAllByRangeForUpdate")
			checkPeriod("FindAllByRangeForUpdate", got)
			return results.slots, results.findAllErr
		},
		findSlotByStartAtAndEndAt: func(_ context.Context, _ repository.Querier, slotStartAt int64, _ int64) (int64, error) {
			*calls = append(*calls, "FindSlotByStartAtAndEndAt")
			return results.counts[slotStartAt], results.findSlotErr
		},
		decrementSlotsByRange: func(_ context.Context, _ repository.Querier, got model.ReservationPeriod) error {
			*calls = append(*calls, "DecrementSlotsByRange")
			checkPeriod("DecrementSlotsByRange", got)
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

func TestLivestreamReservationUsecase_Reserve(t *testing.T) {
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
	}, testReservePeriod)
	logger := &fakeLogger{}
	u := NewLivestreamReservationUsecase(&fakeTxManager{}, livestreamRepo, slotRepo, logger)

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

// 期間の境界値の判定は model.ReservationPeriod のテストで確認している。
// ここでは、予約できない期間の場合に予約枠を触らずにエラーを返すことを確認する。
func TestLivestreamReservationUsecase_Reserve_TimeRange(t *testing.T) {
	tests := []struct {
		name    string
		period  model.ReservationPeriod
		wantErr error
	}{
		{name: "not reservable", period: model.ReservationPeriod{StartAt: 1700870400, EndAt: 1700874000}, wantErr: ErrBadReservationTimeRange},
		{name: "reservable", period: model.ReservationPeriod{StartAt: 1700870400, EndAt: 1700877600}, wantErr: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var slotCalls []string
			slotRepo := newReservationSlotRepositoryForReserve(t, &slotCalls, reservationSlotResults{}, tt.period)
			var livestreamCalls []string
			var created *model.LivestreamModel
			var addedTagIDs []model.TagID
			livestreamRepo := newLivestreamRepositoryForReserve(t, &livestreamCalls, &created, &addedTagIDs, livestreamReserveResults{})
			u := NewLivestreamReservationUsecase(&fakeTxManager{}, livestreamRepo, slotRepo, &fakeLogger{})

			input := testReserveInput
			input.StartAt = tt.period.StartAt
			input.EndAt = tt.period.EndAt
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

func TestLivestreamReservationUsecase_Reserve_Errors(t *testing.T) {
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
			slotRepo := newReservationSlotRepositoryForReserve(t, &slotCalls, tt.slotResults, testReservePeriod)
			var livestreamCalls []string
			var created *model.LivestreamModel
			var addedTagIDs []model.TagID
			livestreamRepo := newLivestreamRepositoryForReserve(t, &livestreamCalls, &created, &addedTagIDs, tt.livestreamResults)
			u := NewLivestreamReservationUsecase(&fakeTxManager{}, livestreamRepo, slotRepo, logger)
			_, err := u.Reserve(context.Background(), 1, testReserveInput)
			if tt.wantUnavailable {
				unavailable, ok := errors.AsType[*ReservationSlotUnavailableError](err)
				if !ok {
					t.Fatalf("err = %v, want *ReservationSlotUnavailableError", err)
				}
				if unavailable.Period != testReservePeriod {
					t.Errorf("unavailable period = %+v, want %+v", unavailable.Period, testReservePeriod)
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

func TestLivestreamReservationUsecase_Reserve_ErrorMessages(t *testing.T) {
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
			slotRepo := newReservationSlotRepositoryForReserve(t, &slotCalls, tt.slotResults, testReservePeriod)
			u := NewLivestreamReservationUsecase(&fakeTxManager{}, &fakeLivestreamRepository{}, slotRepo, &fakeLogger{})
			_, err := u.Reserve(context.Background(), 1, testReserveInput)
			if err == nil || err.Error() != tt.want {
				t.Errorf("err = %v, want %q", err, tt.want)
			}
		})
	}
}
