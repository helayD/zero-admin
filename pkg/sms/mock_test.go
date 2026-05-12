package sms

import (
	"context"
	"testing"
)

func TestMockProviderCodeIsMock(t *testing.T) {
	p := MockProvider{}
	if got := p.Code(); got != MockProviderCode {
		t.Fatalf("MockProvider.Code() = %q, want %q", got, MockProviderCode)
	}
}

func TestMockProviderAlwaysReturnsFixedCode(t *testing.T) {
	p := MockProvider{}

	cases := []struct {
		name   string
		mobile string
		scene  Scene
		cfg    Config
	}{
		{"login default", "13800138001", SceneMemberLogin, Config{}},
		{"login with custom expire", "13912345678", SceneMemberLogin, Config{ExpireSeconds: 120}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := p.Send(context.Background(), tc.mobile, tc.scene, tc.cfg)
			if err != nil {
				t.Fatalf("MockProvider.Send returned error: %v", err)
			}
			if result == nil {
				t.Fatalf("MockProvider.Send returned nil result")
			}
			if result.Code != MockFixedCode {
				t.Fatalf("MockProvider returned code %q, want %q", result.Code, MockFixedCode)
			}
			if result.ProviderCode != MockProviderCode {
				t.Fatalf("MockProvider returned ProviderCode %q, want %q", result.ProviderCode, MockProviderCode)
			}
			expectExpire := tc.cfg.ExpireSeconds
			if expectExpire <= 0 {
				expectExpire = defaultExpireSeconds
			}
			if result.ExpireSeconds != expectExpire {
				t.Fatalf("MockProvider returned ExpireSeconds %d, want %d", result.ExpireSeconds, expectExpire)
			}
		})
	}
}

func TestMockProviderRegisteredOnInit(t *testing.T) {
	provider, ok := ProviderOf(MockProviderCode)
	if !ok {
		t.Fatalf("MockProvider should be auto-registered on init()")
	}
	if provider.Code() != MockProviderCode {
		t.Fatalf("registered provider Code() = %q, want %q", provider.Code(), MockProviderCode)
	}
}
