package memberinfoservicelogic

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestBcryptHashAndCompare(t *testing.T) {
	password := "123456"
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword failed: %v", err)
	}

	// 哈希后的密码应以 $2a$ 或 $2b$ 开头
	if !strings.HasPrefix(string(hashed), "$2a$") && !strings.HasPrefix(string(hashed), "$2b$") {
		t.Fatalf("expected bcrypt prefix, got %q", string(hashed))
	}

	// 正确密码应匹配
	if err := bcrypt.CompareHashAndPassword(hashed, []byte(password)); err != nil {
		t.Fatalf("correct password should match, got: %v", err)
	}

	// 错误密码应不匹配
	if err := bcrypt.CompareHashAndPassword(hashed, []byte("wrong")); err == nil {
		t.Fatal("wrong password should not match")
	}
}

func TestIsBcryptHash(t *testing.T) {
	tests := []struct {
		name     string
		password string
		isBcrypt bool
	}{
		{"bcrypt $2a$", "$2a$10$abcdefghijklmnopqrstuuxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", true},
		{"bcrypt $2b$", "$2b$10$abcdefghijklmnopqrstuuxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", true},
		{"plain text", "123456", false},
		{"empty", "", false},
		{"starts with dollar but not bcrypt", "$other$abc", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := strings.HasPrefix(tc.password, "$2a$") || strings.HasPrefix(tc.password, "$2b$")
			if got != tc.isBcrypt {
				t.Fatalf("isBcrypt(%q) = %v, want %v", tc.password, got, tc.isBcrypt)
			}
		})
	}
}

func TestCreateJwtToken(t *testing.T) {
	secret := "test-secret-key"
	name := "张三"
	mobile := "13800138001"
	var seconds int64 = 86400
	var memberId int64 = 1001

	token, err := createJwtToken(secret, name, mobile, seconds, memberId)
	if err != nil {
		t.Fatalf("createJwtToken failed: %v", err)
	}

	if token == "" {
		t.Fatal("token should not be empty")
	}

	// Token 应由三段 base64 组成，用 . 分隔
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("expected 3 parts in JWT, got %d", len(parts))
	}
}
