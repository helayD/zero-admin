// Story 3.1.1: front-api 验证码登录注册接口逻辑测试。
//
// 用 fake MemberAuthService 替换 svcCtx.MemberAuthService，验证：
//   - 手机号格式校验失败拦截
//   - 验证码格式校验失败拦截
//   - 正常流程透传 RPC 响应
//   - RPC 错误传递

package member

import (
	"context"
	"errors"
	"testing"

	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/rpc/ums/client/memberauthservice"
	"github.com/feihua/zero-admin/rpc/ums/umsclient"
	"google.golang.org/grpc"
)

type fakeMemberAuthService struct {
	sendFn  func(context.Context, *umsclient.SendSmsCodeReq, ...grpc.CallOption) (*umsclient.SendSmsCodeResp, error)
	loginFn func(context.Context, *umsclient.LoginByCodeReq, ...grpc.CallOption) (*umsclient.LoginByCodeResp, error)
}

func (f *fakeMemberAuthService) SendSmsCode(ctx context.Context, in *umsclient.SendSmsCodeReq, opts ...grpc.CallOption) (*umsclient.SendSmsCodeResp, error) {
	if f.sendFn == nil {
		return &umsclient.SendSmsCodeResp{ExpireSeconds: 300}, nil
	}
	return f.sendFn(ctx, in, opts...)
}

func (f *fakeMemberAuthService) LoginByCode(ctx context.Context, in *umsclient.LoginByCodeReq, opts ...grpc.CallOption) (*umsclient.LoginByCodeResp, error) {
	if f.loginFn == nil {
		return &umsclient.LoginByCodeResp{Token: "tk", IsNewUser: false, MemberId: 1001}, nil
	}
	return f.loginFn(ctx, in, opts...)
}

// 编译期断言 fake 实现了 interface
var _ memberauthservice.MemberAuthService = (*fakeMemberAuthService)(nil)

func newAuthSvcCtx(authSvc memberauthservice.MemberAuthService) *svc.ServiceContext {
	return &svc.ServiceContext{MemberAuthService: authSvc}
}

func TestAuthLoginRejectsInvalidMobile(t *testing.T) {
	logic := NewAuthLoginLogic(context.Background(), newAuthSvcCtx(&fakeMemberAuthService{}))
	_, err := logic.AuthLogin(&types.SmsLoginReq{Mobile: "abc", Code: "123456"}, "127.0.0.1")
	if err == nil {
		t.Fatalf("expected error for invalid mobile")
	}
}

func TestAuthLoginRejectsInvalidCode(t *testing.T) {
	logic := NewAuthLoginLogic(context.Background(), newAuthSvcCtx(&fakeMemberAuthService{}))
	_, err := logic.AuthLogin(&types.SmsLoginReq{Mobile: "13800138001", Code: "12"}, "127.0.0.1")
	if err == nil {
		t.Fatalf("expected error for too-short code")
	}
}

func TestAuthLoginAcceptsSixDigitMockCode(t *testing.T) {
	calls := 0
	authSvc := &fakeMemberAuthService{
		loginFn: func(_ context.Context, in *umsclient.LoginByCodeReq, _ ...grpc.CallOption) (*umsclient.LoginByCodeResp, error) {
			calls++
			if in.Mobile != "13800138001" || in.Code != "123456" {
				t.Fatalf("RPC received unexpected req: %+v", in)
			}
			return &umsclient.LoginByCodeResp{Token: "real-token", IsNewUser: true, MemberId: 9999}, nil
		},
	}
	logic := NewAuthLoginLogic(context.Background(), newAuthSvcCtx(authSvc))
	resp, err := logic.AuthLogin(&types.SmsLoginReq{Mobile: "13800138001", Code: "123456", Source: 1}, "127.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected RPC called once, got %d", calls)
	}
	if resp.Code != 0 || resp.Data.Token != "real-token" || !resp.Data.IsNewUser {
		t.Fatalf("unexpected resp: %+v", resp)
	}
}

func TestAuthLoginPropagatesRPCError(t *testing.T) {
	authSvc := &fakeMemberAuthService{
		loginFn: func(_ context.Context, _ *umsclient.LoginByCodeReq, _ ...grpc.CallOption) (*umsclient.LoginByCodeResp, error) {
			return nil, errors.New("验证码错误或已过期")
		},
	}
	logic := NewAuthLoginLogic(context.Background(), newAuthSvcCtx(authSvc))
	_, err := logic.AuthLogin(&types.SmsLoginReq{Mobile: "13800138001", Code: "123456"}, "127.0.0.1")
	if err == nil {
		t.Fatalf("expected RPC error to propagate")
	}
}

func TestAuthSendSmsCodeRejectsInvalidMobile(t *testing.T) {
	logic := NewAuthSendSmsCodeLogic(context.Background(), newAuthSvcCtx(&fakeMemberAuthService{}))
	_, err := logic.AuthSendSmsCode(&types.SendSmsCodeReq{Mobile: "12345"})
	if err == nil {
		t.Fatalf("expected error for invalid mobile")
	}
}

func TestAuthSendSmsCodeReturnsExpireSeconds(t *testing.T) {
	authSvc := &fakeMemberAuthService{
		sendFn: func(_ context.Context, in *umsclient.SendSmsCodeReq, _ ...grpc.CallOption) (*umsclient.SendSmsCodeResp, error) {
			if in.Mobile != "13800138001" {
				t.Fatalf("unexpected mobile: %s", in.Mobile)
			}
			if in.Scene == 0 {
				t.Fatalf("scene default should not be zero after mapping")
			}
			return &umsclient.SendSmsCodeResp{ExpireSeconds: 180}, nil
		},
	}
	logic := NewAuthSendSmsCodeLogic(context.Background(), newAuthSvcCtx(authSvc))
	resp, err := logic.AuthSendSmsCode(&types.SendSmsCodeReq{Mobile: "13800138001"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Code != 0 || resp.Data.ExpireSeconds != 180 {
		t.Fatalf("unexpected resp: %+v", resp)
	}
	// 安全约束：响应体绝不携带验证码原文
	// types.SendSmsCodeData 仅有 ExpireSeconds 字段，编译期已保证；本测试作为契约 lock-in。
}

func TestAuthSendSmsCodePropagatesRPCError(t *testing.T) {
	authSvc := &fakeMemberAuthService{
		sendFn: func(_ context.Context, _ *umsclient.SendSmsCodeReq, _ ...grpc.CallOption) (*umsclient.SendSmsCodeResp, error) {
			return nil, errors.New("验证码请求过于频繁，请稍后再试")
		},
	}
	logic := NewAuthSendSmsCodeLogic(context.Background(), newAuthSvcCtx(authSvc))
	_, err := logic.AuthSendSmsCode(&types.SendSmsCodeReq{Mobile: "13800138001"})
	if err == nil {
		t.Fatalf("expected RPC error to propagate")
	}
}
