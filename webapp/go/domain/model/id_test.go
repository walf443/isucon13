package model

import "testing"

func TestParseID(t *testing.T) {
	tests := []struct {
		s       string
		want    LivestreamID
		wantErr bool
	}{
		{s: "10", want: 10},
		// strconv.Atoi と同じく符号や先頭の 0 も受け付ける (移行前と同じ)
		{s: "+10", want: 10},
		{s: "-1", want: -1},
		{s: "007", want: 7},
		{s: "abc", wantErr: true},
		{s: "", wantErr: true},
		{s: "1.5", wantErr: true},
		{s: "9223372036854775808", wantErr: true},
	}
	for _, tt := range tests {
		got, err := ParseLivestreamID(tt.s)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseLivestreamID(%q) err = %v, wantErr %v", tt.s, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("ParseLivestreamID(%q) = %d, want %d", tt.s, got, tt.want)
		}
	}

	if got, err := ParseLivecommentID("42"); err != nil || got != LivecommentID(42) {
		t.Errorf("ParseLivecommentID(42) = %d, %v", got, err)
	}
}
