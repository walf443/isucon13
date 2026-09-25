package model

import "testing"

func TestIsReservedUsername(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{name: "pipe", want: true},
		{name: "alice", want: false},
		// 完全一致だけを予約済みとする (移行前と同じ)
		{name: "Pipe", want: false},
		{name: "pipeline", want: false},
	}
	for _, tt := range tests {
		if got := IsReservedUsername(tt.name); got != tt.want {
			t.Errorf("IsReservedUsername(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}
