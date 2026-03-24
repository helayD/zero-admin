package member

import (
	"testing"
)

func TestMobileRegexp(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid 138", "13800138001", true},
		{"valid 159", "15912345678", true},
		{"valid 199", "19912345678", true},
		{"too short", "1380013800", false},
		{"too long", "138001380011", false},
		{"invalid prefix 12", "12800138001", false},
		{"invalid prefix 10", "10800138001", false},
		{"empty", "", false},
		{"contains letters", "1380013800a", false},
		{"all zeros", "00000000000", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := mobileRegexp.MatchString(tc.input)
			if got != tc.want {
				t.Fatalf("mobileRegexp.MatchString(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestPasswordLengthValidation(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantOk   bool
	}{
		{"6 chars", "123456", true},
		{"8 chars", "12345678", true},
		{"5 chars", "12345", false},
		{"empty", "", false},
		{"1 char", "a", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ok := len(tc.password) >= 6
			if ok != tc.wantOk {
				t.Fatalf("len(%q) >= 6 = %v, want %v", tc.password, ok, tc.wantOk)
			}
		})
	}
}

func TestPasswordConfirmValidation(t *testing.T) {
	tests := []struct {
		name            string
		password        string
		confirmPassword string
		wantMatch       bool
	}{
		{"match", "123456", "123456", true},
		{"mismatch", "123456", "654321", false},
		{"empty confirm", "123456", "", false},
		{"both empty", "", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.password == tc.confirmPassword
			if got != tc.wantMatch {
				t.Fatalf("password match = %v, want %v", got, tc.wantMatch)
			}
		})
	}
}
