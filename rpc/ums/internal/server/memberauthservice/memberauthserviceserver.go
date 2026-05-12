// Hand-written gRPC service implementation for MemberAuthService.
// Story 3.1.1 引入：手机号+验证码合并登录注册接口。
package server

import (
	"context"

	memberinfoservicelogic "github.com/feihua/zero-admin/rpc/ums/internal/logic/memberinfoservice"
	"github.com/feihua/zero-admin/rpc/ums/internal/svc"
	"github.com/feihua/zero-admin/rpc/ums/umsclient"
)

// MemberAuthServiceServer 实现 umsclient.MemberAuthServiceServer。
//
// 注意 logic 复用 memberinfoservicelogic 包，因为：
//   - createJwtToken / grantDailyLoginLotteryTimes / member id 序列等公共能力
//     已经在该包中实现，避免重复
//   - gRPC 服务边界与 Go 包边界无需一一对应
type MemberAuthServiceServer struct {
	svcCtx *svc.ServiceContext
	umsclient.UnimplementedMemberAuthServiceServer
}

func NewMemberAuthServiceServer(svcCtx *svc.ServiceContext) *MemberAuthServiceServer {
	return &MemberAuthServiceServer{svcCtx: svcCtx}
}

// SendSmsCode 发送短信验证码（手机号+场景）
func (s *MemberAuthServiceServer) SendSmsCode(ctx context.Context, in *umsclient.SendSmsCodeReq) (*umsclient.SendSmsCodeResp, error) {
	l := memberinfoservicelogic.NewSendSmsCodeLogic(ctx, s.svcCtx)
	return l.SendSmsCode(in)
}

// LoginByCode 验证码登录注册合并接口（已注册→登录；未注册→自动建号）
func (s *MemberAuthServiceServer) LoginByCode(ctx context.Context, in *umsclient.LoginByCodeReq) (*umsclient.LoginByCodeResp, error) {
	l := memberinfoservicelogic.NewLoginByCodeLogic(ctx, s.svcCtx)
	return l.LoginByCode(in)
}
