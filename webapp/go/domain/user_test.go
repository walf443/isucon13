package domain

import "testing"

func TestFindReservedUsername(t *testing.T) {
	tests := []struct {
		name         string
		wantReserved string
		wantOK       bool
	}{
		{name: "pipe", wantReserved: "pipe", wantOK: true},
		{name: "alice"},
		// 完全一致だけを予約済みとする (移行前と同じ)
		{name: "Pipe"},
		{name: "pipeline"},
	}
	for _, tt := range tests {
		reserved, ok := FindReservedUsername(tt.name)
		if reserved != tt.wantReserved || ok != tt.wantOK {
			t.Errorf("FindReservedUsername(%q) = %q, %v, want %q, %v", tt.name, reserved, ok, tt.wantReserved, tt.wantOK)
		}
	}
}
