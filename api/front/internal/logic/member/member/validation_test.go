// Story 3.1.1: 旧的密码登录/注册接口已下线，validation_test 仅保留
// mobileRegexp 与 smsCodeRegexp 的通用校验测试。

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

func TestSmsCodeRegexp(t *testing.T) {
	// AC-12: 验证码硬约束 6 位纯数字（mock 与真实 provider 统一）。
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"mock 6 digits 123456", "123456", true},
		{"real provider 6 digits", "456789", true},
		{"4 digits rejected", "1234", false},
		{"5 digits rejected", "12345", false},
		{"too short 3", "123", false},
		{"too long 7", "1234567", false},
		{"empty", "", false},
		{"letters", "12a456", false},
		{"with space", "123 456", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := smsCodeRegexp.MatchString(tc.input)
			if got != tc.want {
				t.Fatalf("smsCodeRegexp.MatchString(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}
