package sms

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type stubResolver struct {
	cfg Config
	err error
}

func (s stubResolver) Resolve(_ context.Context, _ Scene) (Config, error) {
	return s.cfg, s.err
}

type stubProvider struct {
	code        string
	result      *SendResult
	err         error
	calledTimes int
}

func (s *stubProvider) Code() string { return s.code }
func (s *stubProvider) Send(_ context.Context, _ string, _ Scene, _ Config) (*SendResult, error) {
	s.calledTimes++
	return s.result, s.err
}

func TestDefaultSenderRoutesToConfiguredProvider(t *testing.T) {
	resetRegistryForTest()
	defer func() {
		resetRegistryForTest()
		// 恢复 mock provider，避免污染其他测试
		RegisterProvider(MockProvider{})
	}()

	provider := &stubProvider{
		code:   "stub",
		result: &SendResult{Code: "999999", ExpireSeconds: 300, ProviderCode: "stub"},
	}
	RegisterProvider(provider)

	resolver := stubResolver{cfg: Config{ProviderCode: "stub", ExpireSeconds: 300}}
	sender := NewSender(resolver)

	got, err := sender.Send(context.Background(), "13800138001", SceneMemberLogin)
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if got.Code != "999999" {
		t.Fatalf("Send returned Code %q, want 999999", got.Code)
	}
	if provider.calledTimes != 1 {
		t.Fatalf("provider called %d times, want 1", provider.calledTimes)
	}
}

func TestDefaultSenderUnknownProviderReturnsError(t *testing.T) {
	resetRegistryForTest()
	defer func() {
		resetRegistryForTest()
		RegisterProvider(MockProvider{})
	}()

	resolver := stubResolver{cfg: Config{ProviderCode: "non-existent"}}
	sender := NewSender(resolver)

	_, err := sender.Send(context.Background(), "13800138001", SceneMemberLogin)
	if err == nil {
		t.Fatalf("Send should return error when provider unregistered")
	}
	if !strings.Contains(err.Error(), "non-existent") {
		t.Fatalf("error should mention provider code, got: %v", err)
	}
}

func TestDefaultSenderResolverErrorWrapped(t *testing.T) {
	baseErr := errors.New("sys-rpc unreachable")
	sender := NewSender(stubResolver{err: baseErr})

	_, err := sender.Send(context.Background(), "13800138001", SceneMemberLogin)
	if err == nil {
		t.Fatalf("Send should return error when resolver fails")
	}
	if !errors.Is(err, baseErr) {
		t.Fatalf("error chain should include resolver error, got: %v", err)
	}
}

func TestDefaultSenderEmptyMobileReturnsError(t *testing.T) {
	sender := NewSender(stubResolver{cfg: Config{ProviderCode: MockProviderCode}})
	if _, err := sender.Send(context.Background(), "  ", SceneMemberLogin); err == nil {
		t.Fatalf("Send with empty mobile should return error")
	}
}

func TestDefaultSenderNilResolverReturnsError(t *testing.T) {
	sender := NewSender(nil)
	if _, err := sender.Send(context.Background(), "13800138001", SceneMemberLogin); err == nil {
		t.Fatalf("Send with nil resolver should return error")
	}
}

func TestDefaultSenderEmptyProviderCodeReturnsError(t *testing.T) {
	sender := NewSender(stubResolver{cfg: Config{ProviderCode: "  "}})
	if _, err := sender.Send(context.Background(), "13800138001", SceneMemberLogin); err == nil {
		t.Fatalf("Send with empty ProviderCode should return error")
	}
}
