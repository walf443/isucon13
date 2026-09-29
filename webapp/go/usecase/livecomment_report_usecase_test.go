package usecase

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// newReportRepositoryFindingAll はライブ配信 1 への報告として models (失敗させる場合は err) を返す fakeLivecommentReportRepository を返す。
// 呼ばれた回数を calls に数える。
func newReportRepositoryFindingAll(t *testing.T, calls *int, models []*domain.LivecommentReport, err error) *fakeLivecommentReportRepository {
	return &fakeLivecommentReportRepository{
		findAllByLivestreamID: func(_ context.Context, _ repository.Querier, livestreamID domain.LivestreamID) ([]*domain.LivecommentReport, error) {
			*calls++
			if livestreamID != 1 {
				t.Errorf("report livestream id = %d, want 1", livestreamID)
			}
			return models, err
		},
	}
}

func TestLivecommentReportUsecase_FindAllByLivestreamID(t *testing.T) {
	f := testLivestreamFixture()
	var reportCalls int
	reportRepo := newReportRepositoryFindingAll(t, &reportCalls, []*domain.LivecommentReport{testReportModel2, testReportModel1}, nil)
	u := NewLivecommentReportUsecase(&fakeTxManager{}, newLivestreamRepositoryFindingByID(t, 1, testLivestreamModel1, nil), nil, reportRepo, f.reportFiller(testLivecommentModel1, testLivecommentModel2))

	// ライブ配信 1 の配信者 (42) として取得する
	got, err := u.FindAllByLivestreamID(context.Background(), 42, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// repository が返した順序のまま
	want := []*domain.LivecommentReportDetail{
		f.report(testReportModel2, testLivecommentModel2, testLivestreamModel1),
		f.report(testReportModel1, testLivecommentModel1, testLivestreamModel1),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestLivecommentReportUsecase_FindAllByLivestreamID_Errors(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name string
		// userID は取得するユーザ
		userID domain.UserID
		// livestream, livestreamErr はライブ配信 1 を引いた結果
		livestream    *domain.Livestream
		livestreamErr error
		// reports, reportsErr は報告の取得の結果
		reports         []*domain.LivecommentReport
		reportsErr      error
		wantErr         error
		wantMsg         string
		wantReportCalls int
	}{
		{
			name:       "not the owner",
			userID:     43,
			livestream: testLivestreamModel1,
			wantErr:    ErrNotLivestreamOwner,
			wantMsg:    ErrNotLivestreamOwner.Error(),
		},
		{
			// ライブ配信が無い場合は 404 用のエラーには変換しない (移行前と同じく 500)
			name:          "livestream not found",
			userID:        42,
			livestreamErr: repository.ErrNotFound,
			wantErr:       repository.ErrNotFound,
			wantMsg:       "failed to get livestream: not found",
		},
		{
			name:            "report repository error",
			userID:          42,
			livestream:      testLivestreamModel1,
			reportsErr:      boom,
			wantErr:         boom,
			wantMsg:         "failed to get livecomment reports: boom",
			wantReportCalls: 1,
		},
		{
			// 報告されたライブコメント (52) は引けないので組み立てに失敗する
			name:            "fill fails",
			userID:          42,
			livestream:      testLivestreamModel1,
			reports:         []*domain.LivecommentReport{{ID: 9, UserID: 42, LivestreamID: 1, LivecommentID: 52}},
			wantErr:         repository.ErrNotFound,
			wantMsg:         "failed to get livecomment reports: failed to get livecomment of livecomment report 9: not found",
			wantReportCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reportCalls int
			reportRepo := newReportRepositoryFindingAll(t, &reportCalls, tt.reports, tt.reportsErr)
			livestreamRepo := newLivestreamRepositoryFindingByID(t, 1, tt.livestream, tt.livestreamErr)
			u := NewLivecommentReportUsecase(&fakeTxManager{}, livestreamRepo, nil, reportRepo, testLivestreamFixture().reportFiller(testLivecommentModel1))
			_, err := u.FindAllByLivestreamID(context.Background(), tt.userID, 1)
			if !errors.Is(err, tt.wantErr) || err.Error() != tt.wantMsg {
				t.Errorf("err = %v, want %q", err, tt.wantMsg)
			}
			if errors.Is(err, ErrLivestreamNotFound) {
				t.Errorf("err = %v, should not be ErrLivestreamNotFound", err)
			}
			if reportCalls != tt.wantReportCalls {
				t.Errorf("report calls = %d, want %d", reportCalls, tt.wantReportCalls)
			}
		})
	}
}

// newReportRepositoryForCreate は Create で呼ばれるメソッドを、呼ばれた順に calls へ記録する fakeLivecommentReportRepository を返す。
// 登録した報告は created に取り出し、ID 7 で読み直した報告として report を返す。
func newReportRepositoryForCreate(t *testing.T, calls *[]string, created **domain.LivecommentReport, report *domain.LivecommentReport, createErr, findErr error) *fakeLivecommentReportRepository {
	return &fakeLivecommentReportRepository{
		create: func(_ context.Context, _ repository.Querier, r *domain.LivecommentReport) (domain.LivecommentReportID, error) {
			*calls = append(*calls, "Create")
			*created = r
			return 7, createErr
		},
		findByID: func(_ context.Context, _ repository.Querier, id domain.LivecommentReportID) (*domain.LivecommentReport, error) {
			*calls = append(*calls, "FindByID")
			if id != 7 {
				t.Errorf("re-read id = %d, want 7", id)
			}
			return report, findErr
		},
	}
}

// newLivecommentRepositoryForReport はライブコメント 50 の存在確認の結果として livecomment (失敗させる場合は err) を返す fakeLivecommentRepository を返す。
// 呼ばれた回数を calls に数える。
func newLivecommentRepositoryForReport(t *testing.T, calls *int, livecomment *domain.Livecomment, err error) *fakeLivecommentRepository {
	return &fakeLivecommentRepository{
		findByID: func(_ context.Context, _ repository.Querier, id domain.LivecommentID) (*domain.Livecomment, error) {
			*calls++
			if id != 50 {
				t.Errorf("livecomment id = %d, want 50", id)
			}
			return livecomment, err
		},
	}
}

// testCreatedReportModel は登録した報告 (ID 7) を読み直した結果。
var testCreatedReportModel = &domain.LivecommentReport{ID: 7, UserID: 43, LivestreamID: 1, LivecommentID: 50, CreatedAt: 1700000000}

func TestLivecommentReportUsecase_Create(t *testing.T) {
	f := testLivestreamFixture()
	var livecommentFinds int
	livecommentRepo := newLivecommentRepositoryForReport(t, &livecommentFinds, testLivecommentModel1, nil)
	var reportCalls []string
	var created *domain.LivecommentReport
	reportRepo := newReportRepositoryForCreate(t, &reportCalls, &created, testCreatedReportModel, nil, nil)
	livestreamRepo := newLivestreamRepositoryFindingByID(t, 1, testLivestreamModel1, nil)
	u := NewLivecommentReportUsecase(&fakeTxManager{}, livestreamRepo, livecommentRepo, reportRepo, f.reportFiller(testLivecommentModel1)).(*livecommentReportUsecase)
	u.now = func() time.Time { return time.Unix(1700000000, 0) }

	got, err := u.Create(context.Background(), 43, 1, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := f.report(testCreatedReportModel, testLivecommentModel1, testLivestreamModel1); !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
	if livecommentFinds != 1 {
		t.Errorf("livecomment finds = %d, want 1", livecommentFinds)
	}
	wantCreated := domain.LivecommentReport{UserID: 43, LivestreamID: 1, LivecommentID: 50, CreatedAt: 1700000000}
	if created == nil || *created != wantCreated {
		t.Errorf("created = %+v, want %+v", created, wantCreated)
	}
	if want := []string{"Create", "FindByID"}; !slices.Equal(reportCalls, want) {
		t.Errorf("report calls = %v, want %v", reportCalls, want)
	}
}

func TestLivecommentReportUsecase_Create_Errors(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name string
		// livestream, livestreamErr はライブ配信 1 を引いた結果
		livestream    *domain.Livestream
		livestreamErr error
		// livecomment, livecommentErr はライブコメントの存在確認の結果
		livecomment    *domain.Livecomment
		livecommentErr error
		// reportCreateErr, reportFindErr は報告の登録・取り直しが返すエラー
		reportCreateErr error
		reportFindErr   error
		// modify はデータを欠けさせる
		modify          func(f *livestreamFixture)
		wantErr         error
		wantMsg         string
		wantReportCalls []string
	}{
		{name: "livestream not found", livestreamErr: repository.ErrNotFound, wantErr: ErrLivestreamNotFound, wantMsg: ErrLivestreamNotFound.Error()},
		{name: "livestream repository error", livestreamErr: boom, wantErr: boom, wantMsg: "failed to get livestream: boom"},
		{name: "livecomment not found", livestream: testLivestreamModel1, livecommentErr: repository.ErrNotFound, wantErr: ErrLivecommentNotFound, wantMsg: ErrLivecommentNotFound.Error()},
		{name: "livecomment repository error", livestream: testLivestreamModel1, livecommentErr: boom, wantErr: boom, wantMsg: "failed to get livecomment: boom"},
		{
			name:            "create fails",
			livestream:      testLivestreamModel1,
			livecomment:     testLivecommentModel1,
			reportCreateErr: boom,
			wantErr:         boom,
			wantMsg:         "failed to insert livecomment report: boom",
			wantReportCalls: []string{"Create"},
		},
		{
			name:            "re-read fails",
			livestream:      testLivestreamModel1,
			livecomment:     testLivecommentModel1,
			reportFindErr:   boom,
			wantErr:         boom,
			wantMsg:         "failed to fill livecomment report: boom",
			wantReportCalls: []string{"Create", "FindByID"},
		},
		{
			name:            "fill fails",
			livestream:      testLivestreamModel1,
			livecomment:     testLivecommentModel1,
			modify:          func(f *livestreamFixture) { delete(f.users, 43) },
			wantErr:         repository.ErrNotFound,
			wantMsg:         "failed to fill livecomment report: failed to get reporter of livecomment report 7: not found",
			wantReportCalls: []string{"Create", "FindByID"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := testLivestreamFixture()
			if tt.modify != nil {
				tt.modify(f)
			}
			var reportCalls []string
			var created *domain.LivecommentReport
			reportRepo := newReportRepositoryForCreate(t, &reportCalls, &created, testCreatedReportModel, tt.reportCreateErr, tt.reportFindErr)
			var livecommentFinds int
			livecommentRepo := newLivecommentRepositoryForReport(t, &livecommentFinds, tt.livecomment, tt.livecommentErr)
			livestreamRepo := newLivestreamRepositoryFindingByID(t, 1, tt.livestream, tt.livestreamErr)
			u := NewLivecommentReportUsecase(&fakeTxManager{}, livestreamRepo, livecommentRepo, reportRepo, f.reportFiller(testLivecommentModel1))
			_, err := u.Create(context.Background(), 43, 1, 50)
			if !errors.Is(err, tt.wantErr) || err.Error() != tt.wantMsg {
				t.Errorf("err = %v, want %q", err, tt.wantMsg)
			}
			// ライブ配信やライブコメントが無い場合は報告を登録しない
			if !slices.Equal(reportCalls, tt.wantReportCalls) {
				t.Errorf("report calls = %v, want %v", reportCalls, tt.wantReportCalls)
			}
		})
	}
}
