package usecase

import (
	"errors"
	"fmt"
	"testing"

	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

func TestAsMissingDetail(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name string
		err  error
		// wantMissing は errMissingDetail として判定されることを期待するかどうか
		wantMissing bool
		// wantErr は errors.Is で判定できることを期待する元のエラー (無ければ nil)
		wantErr error
	}{
		{name: "not found", err: repository.ErrNotFound, wantMissing: true},
		{name: "wrapped not found", err: fmt.Errorf("wrapped: %w", repository.ErrNotFound), wantMissing: true},
		{name: "other error", err: boom, wantErr: boom},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := asMissingDetail(tt.err)
			// メッセージは変えない
			if got.Error() != tt.err.Error() {
				t.Errorf("message = %q, want %q", got.Error(), tt.err.Error())
			}
			if errors.Is(got, errMissingDetail) != tt.wantMissing {
				t.Errorf("errors.Is(errMissingDetail) = %v, want %v", !tt.wantMissing, tt.wantMissing)
			}
			// 付随するデータの欠損は ErrNotFound (主となるデータの不在) として判定されない
			if tt.wantMissing && errors.Is(got, repository.ErrNotFound) {
				t.Errorf("err = %v, should not be ErrNotFound", got)
			}
			if tt.wantErr != nil && !errors.Is(got, tt.wantErr) {
				t.Errorf("err = %v, want %v", got, tt.wantErr)
			}
		})
	}
}
