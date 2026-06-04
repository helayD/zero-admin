package sms

import (
	"context"
	"errors"
	"strings"
	"testing"

	dysmsapi "github.com/alibabacloud-go/dysmsapi-20170525/v5/client"
	"github.com/alibabacloud-go/tea/tea"
)

func TestAliyunCredentialEnvNames(t *testing.T) {
	cases := []struct {
		name       string
		ref        string
		wantIDEnv  string
		wantSecEnv string
	}{
		{
			name:       "default",
			ref:        "",
			wantIDEnv:  aliyunDefaultAccessKeyIDEnv,
			wantSecEnv: aliyunDefaultSecretEnv,
		},
		{
			name:       "prefix",
			ref:        "ref:env:ALIYUN_SMS",
			wantIDEnv:  "ALIYUN_SMS_ACCESS_KEY_ID",
			wantSecEnv: "ALIYUN_SMS_ACCESS_KEY_SECRET",
		},
		{
			name:       "explicit",
			ref:        "ref:env:ALIYUN_ID,ALIYUN_SECRET",
			wantIDEnv:  "ALIYUN_ID",
			wantSecEnv: "ALIYUN_SECRET",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotID, gotSecret, err := aliyunCredentialEnvNames(tc.ref)
			if err != nil {
				t.Fatalf("aliyunCredentialEnvNames returned error: %v", err)
			}
			if gotID != tc.wantIDEnv || gotSecret != tc.wantSecEnv {
				t.Fatalf("got %s/%s, want %s/%s", gotID, gotSecret, tc.wantIDEnv, tc.wantSecEnv)
			}
		})
	}
}

func TestAliyunCredentialEnvNamesRejectsPlainRef(t *testing.T) {
	_, _, err := aliyunCredentialEnvNames("credential://plain")
	if err == nil {
		t.Fatal("expected plain credential ref to fail")
	}
}

func TestResolveAliyunCredentialReadsEnvironment(t *testing.T) {
	t.Setenv("TEST_ALIYUN_ACCESS_KEY_ID", "ak")
	t.Setenv("TEST_ALIYUN_ACCESS_KEY_SECRET", "sk")

	gotID, gotSecret, err := resolveAliyunCredential(Config{CredentialRef: "ref:env:TEST_ALIYUN"})
	if err != nil {
		t.Fatalf("resolveAliyunCredential returned error: %v", err)
	}
	if gotID != "ak" || gotSecret != "sk" {
		t.Fatalf("got %s/%s, want ak/sk", gotID, gotSecret)
	}
}

func TestAliyunProviderSendsExpectedRequest(t *testing.T) {
	t.Setenv("TEST_ALIYUN_ACCESS_KEY_ID", "ak")
	t.Setenv("TEST_ALIYUN_ACCESS_KEY_SECRET", "sk")

	oldSend := aliyunSendSms
	defer func() { aliyunSendSms = oldSend }()

	var gotRuntime aliyunRuntimeConfig
	var gotReq *dysmsapi.SendSmsRequest
	aliyunSendSms = func(_ context.Context, runtimeCfg aliyunRuntimeConfig, req *dysmsapi.SendSmsRequest) (*dysmsapi.SendSmsResponse, error) {
		gotRuntime = runtimeCfg
		gotReq = req
		return &dysmsapi.SendSmsResponse{
			Body: &dysmsapi.SendSmsResponseBody{
				Code:    tea.String("OK"),
				Message: tea.String("OK"),
			},
		}, nil
	}

	result, err := (AliyunProvider{}).Send(context.Background(), "13800138001", SceneMemberLogin, Config{
		SignName:       "杭州山河集信息科技",
		TemplateCode:   "SMS_335160513",
		CredentialRef:  "ref:env:TEST_ALIYUN",
		Endpoint:       "dysmsapi.test.aliyuncs.com",
		TimeoutSeconds: 7,
		ExpireSeconds:  180,
	})
	if err != nil {
		t.Fatalf("AliyunProvider.Send returned error: %v", err)
	}
	if result.ProviderCode != AliyunProviderCode {
		t.Fatalf("ProviderCode = %s, want %s", result.ProviderCode, AliyunProviderCode)
	}
	if len(result.Code) != 6 {
		t.Fatalf("code length = %d, want 6", len(result.Code))
	}
	if result.ExpireSeconds != 180 {
		t.Fatalf("ExpireSeconds = %d, want 180", result.ExpireSeconds)
	}
	if gotRuntime.AccessKeyID != "ak" || gotRuntime.AccessKeySecret != "sk" {
		t.Fatalf("credential = %s/%s, want ak/sk", gotRuntime.AccessKeyID, gotRuntime.AccessKeySecret)
	}
	if gotRuntime.Endpoint != "dysmsapi.test.aliyuncs.com" {
		t.Fatalf("endpoint = %s", gotRuntime.Endpoint)
	}
	if gotReq == nil {
		t.Fatal("request should not be nil")
	}
	if tea.StringValue(gotReq.PhoneNumbers) != "13800138001" {
		t.Fatalf("PhoneNumbers = %s", tea.StringValue(gotReq.PhoneNumbers))
	}
	if tea.StringValue(gotReq.SignName) != "杭州山河集信息科技" {
		t.Fatalf("SignName = %s", tea.StringValue(gotReq.SignName))
	}
	if tea.StringValue(gotReq.TemplateCode) != "SMS_335160513" {
		t.Fatalf("TemplateCode = %s", tea.StringValue(gotReq.TemplateCode))
	}
	if !strings.Contains(tea.StringValue(gotReq.TemplateParam), `"code"`) {
		t.Fatalf("TemplateParam should include code, got %s", tea.StringValue(gotReq.TemplateParam))
	}
	if gotRuntime.TimeoutSeconds != 7 {
		t.Fatalf("timeout = %d, want 7", gotRuntime.TimeoutSeconds)
	}
}

func TestAliyunProviderPrefersConfigCredentials(t *testing.T) {
	oldSend := aliyunSendSms
	defer func() { aliyunSendSms = oldSend }()

	var gotRuntime aliyunRuntimeConfig
	aliyunSendSms = func(_ context.Context, runtimeCfg aliyunRuntimeConfig, _ *dysmsapi.SendSmsRequest) (*dysmsapi.SendSmsResponse, error) {
		gotRuntime = runtimeCfg
		return &dysmsapi.SendSmsResponse{
			Body: &dysmsapi.SendSmsResponseBody{
				Code:    tea.String("OK"),
				Message: tea.String("OK"),
			},
		}, nil
	}

	_, err := (AliyunProvider{}).Send(context.Background(), "13800138001", SceneMemberLogin, Config{
		SignName:        "杭州山河集信息科技",
		TemplateCode:    "SMS_335160513",
		AccessKeyID:     "db-ak",
		AccessKeySecret: "db-sk",
		CredentialRef:   "ref:env:MISSING_ENV",
	})
	if err != nil {
		t.Fatalf("AliyunProvider.Send returned error: %v", err)
	}
	if gotRuntime.AccessKeyID != "db-ak" || gotRuntime.AccessKeySecret != "db-sk" {
		t.Fatalf("runtime credential = %s/%s, want db-ak/db-sk", gotRuntime.AccessKeyID, gotRuntime.AccessKeySecret)
	}
	if gotRuntime.Endpoint != aliyunDefaultEndpoint {
		t.Fatalf("default endpoint = %s, want %s", gotRuntime.Endpoint, aliyunDefaultEndpoint)
	}
}

func TestAliyunProviderReturnsGatewayFailure(t *testing.T) {
	t.Setenv("TEST_ALIYUN_ACCESS_KEY_ID", "ak")
	t.Setenv("TEST_ALIYUN_ACCESS_KEY_SECRET", "sk")

	oldSend := aliyunSendSms
	defer func() { aliyunSendSms = oldSend }()
	aliyunSendSms = func(context.Context, aliyunRuntimeConfig, *dysmsapi.SendSmsRequest) (*dysmsapi.SendSmsResponse, error) {
		return &dysmsapi.SendSmsResponse{
			Body: &dysmsapi.SendSmsResponseBody{
				Code:    tea.String("isv.BUSINESS_LIMIT_CONTROL"),
				Message: tea.String("触发业务流控"),
			},
		}, nil
	}

	_, err := (AliyunProvider{}).Send(context.Background(), "13800138001", SceneMemberLogin, Config{
		SignName:      "杭州山河集信息科技",
		TemplateCode:  "SMS_335160513",
		CredentialRef: "ref:env:TEST_ALIYUN",
	})
	if err == nil {
		t.Fatal("expected gateway failure")
	}
	if !strings.Contains(err.Error(), "BUSINESS_LIMIT_CONTROL") {
		t.Fatalf("error should include gateway code, got %v", err)
	}
}

func TestAliyunProviderWrapsSendError(t *testing.T) {
	t.Setenv("TEST_ALIYUN_ACCESS_KEY_ID", "ak")
	t.Setenv("TEST_ALIYUN_ACCESS_KEY_SECRET", "sk")

	oldSend := aliyunSendSms
	defer func() { aliyunSendSms = oldSend }()
	baseErr := errors.New("network down")
	aliyunSendSms = func(context.Context, aliyunRuntimeConfig, *dysmsapi.SendSmsRequest) (*dysmsapi.SendSmsResponse, error) {
		return nil, baseErr
	}

	_, err := (AliyunProvider{}).Send(context.Background(), "13800138001", SceneMemberLogin, Config{
		SignName:      "杭州山河集信息科技",
		TemplateCode:  "SMS_335160513",
		CredentialRef: "ref:env:TEST_ALIYUN",
	})
	if !errors.Is(err, baseErr) {
		t.Fatalf("error chain should include send error, got %v", err)
	}
}

func TestAliyunProviderRegisteredOnInit(t *testing.T) {
	provider, ok := ProviderOf(AliyunProviderCode)
	if !ok {
		t.Fatalf("AliyunProvider should be auto-registered on init()")
	}
	if provider.Code() != AliyunProviderCode {
		t.Fatalf("registered provider Code() = %q, want %q", provider.Code(), AliyunProviderCode)
	}
}
