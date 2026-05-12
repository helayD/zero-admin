// Story 3.1.1: 注册手写的 MemberAuthService（手机号+验证码合并登录注册）。
// 与 sys-rpc 的 RegisterExtraServices 模式保持一致，避免污染 ums.go 主入口。

package server

import (
	memberauthserviceServer "github.com/feihua/zero-admin/rpc/ums/internal/server/memberauthservice"
	"github.com/feihua/zero-admin/rpc/ums/internal/svc"
	"github.com/feihua/zero-admin/rpc/ums/umsclient"
	"google.golang.org/grpc"
)

// RegisterExtraServices 注册手写的额外 gRPC 服务到 grpc.Server。
// 主入口 rpc/ums/ums.go 在初始化完所有由 protoc 生成的 service 之后调用本函数。
func RegisterExtraServices(grpcServer *grpc.Server, svcCtx *svc.ServiceContext) {
	umsclient.RegisterMemberAuthServiceServer(grpcServer, memberauthserviceServer.NewMemberAuthServiceServer(svcCtx))
}
