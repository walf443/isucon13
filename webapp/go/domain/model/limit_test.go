package model

import (
	"errors"
	"testing"
)

func TestParseLimit(t *testing.T) {
	tests := []struct {
		s          string
		want       Limit
		wantRange  bool
		wantNotInt bool
	}{
		{s: "1", want: 1},
		{s: "100", want: 100},
		{s: "+10", want: 10},
		// 0 件を取得する意味は無いので範囲外
		{s: "0", wantRange: true},
		{s: "-1", wantRange: true},
		{s: "101", wantRange: true},
		{s: "abc", wantNotInt: true},
		{s: "", wantNotInt: true},
		{s: "9223372036854775808", wantNotInt: true},
	}
	for _, tt := range tests {
		got, err := ParseLimit(tt.s, 100)
		switch {
		case tt.wantRange:
			if !errors.Is(err, ErrLimitOutOfRange) {
				t.Errorf("ParseLimit(%q) err = %v, want ErrLimitOutOfRange", tt.s, err)
			}
		case tt.wantNotInt:
			if err == nil || errors.Is(err, ErrLimitOutOfRange) {
				t.Errorf("ParseLimit(%q) err = %v, want a parse error", tt.s, err)
			}
		default:
			if err != nil || got != tt.want {
				t.Errorf("ParseLimit(%q) = %d, %v, want %d", tt.s, got, err, tt.want)
			}
		}
	}
}
