// Hand-written go-zero zrpc client wrapper for MemberAuthService.
// Story 3.1.1 引入；与 client/memberinfoservice 同等用法。

package memberauthservice

import (
	"context"

	"github.com/feihua/zero-admin/rpc/ums/umsclient"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type (
	SendSmsCodeReq  = umsclient.SendSmsCodeReq
	SendSmsCodeResp = umsclient.SendSmsCodeResp
	LoginByCodeReq  = umsclient.LoginByCodeReq
	LoginByCodeResp = umsclient.LoginByCodeResp

	MemberAuthService interface {
		SendSmsCode(ctx context.Context, in *SendSmsCodeReq, opts ...grpc.CallOption) (*SendSmsCodeResp, error)
		LoginByCode(ctx context.Context, in *LoginByCodeReq, opts ...grpc.CallOption) (*LoginByCodeResp, error)
	}

	defaultMemberAuthService struct {
		cli zrpc.Client
	}
)

func NewMemberAuthService(cli zrpc.Client) MemberAuthService {
	return &defaultMemberAuthService{cli: cli}
}

func (m *defaultMemberAuthService) SendSmsCode(ctx context.Context, in *SendSmsCodeReq, opts ...grpc.CallOption) (*SendSmsCodeResp, error) {
	client := umsclient.NewMemberAuthServiceClient(m.cli.Conn())
	return client.SendSmsCode(ctx, in, opts...)
}

func (m *defaultMemberAuthService) LoginByCode(ctx context.Context, in *LoginByCodeReq, opts ...grpc.CallOption) (*LoginByCodeResp, error) {
	client := umsclient.NewMemberAuthServiceClient(m.cli.Conn())
	return client.LoginByCode(ctx, in, opts...)
}
