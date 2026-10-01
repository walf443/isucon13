package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestFindReservedUsername(t *testing.T) {
	tests := []struct {
		name         Username
		wantReserved Username
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

func TestParseUsername(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{name: "lowercase letters", input: "alice", valid: true},
		{name: "keeps the case", input: "AliCe", valid: true},
		{name: "letters and digits", input: "alice01", valid: true},
		{name: "digits only", input: "12345", valid: true},
		{name: "hyphen in the middle", input: "alice-bob", valid: true},
		{name: "one character", input: "a", valid: true},
		{name: "63 characters", input: strings.Repeat("a", 63), valid: true},
		{name: "empty", input: ""},
		{name: "64 characters", input: strings.Repeat("a", 64)},
		{name: "starts with a hyphen (looks like a command option)", input: "-alice"},
		{name: "ends with a hyphen", input: "alice-"},
		{name: "only a hyphen", input: "-"},
		{name: "dot would make a deeper subdomain", input: "a.b"},
		{name: "leading dot", input: ".alice"},
		{name: "underscore", input: "alice_bob"},
		{name: "space", input: "alice bob"},
		{name: "trailing newline", input: "alice\n"},
		{name: "slash", input: "alice/bob"},
		{name: "at sign", input: "alice@bob"},
		{name: "wildcard", input: "*"},
		{name: "NUL byte", input: "alice\x00"},
		{name: "non-ASCII letters", input: "ありす"},
		{name: "non-ASCII lookalike letter", input: "alicé"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseUsername(tt.input)
			if tt.valid {
				if err != nil || got != Username(tt.input) {
					t.Errorf("ParseUsername(%q) = %q, %v, want it to be accepted unchanged", tt.input, got, err)
				}
				return
			}
			if !errors.Is(err, ErrInvalidUsername) || got != "" {
				t.Errorf("ParseUsername(%q) = %q, %v, want ErrInvalidUsername", tt.input, got, err)
			}
		})
	}
}
