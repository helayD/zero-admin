// Story 3.1.1: front-api SendSmsCode logic
//
// 前置校验:
//   - 手机号格式: ^1[3-9]\d{9}$（复用 mobileRegexp）
//   - 场景默认 1 (member_login)
//
// 业务委托给 ums-rpc.MemberAuthService.SendSmsCode 完成实际下发与 Redis 写入。

package member

import (
	"context"

	"github.com/feihua/zero-admin/pkg/errorx"
	"github.com/feihua/zero-admin/rpc/ums/umsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"google.golang.org/grpc/status"

	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AuthSendSmsCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAuthSendSmsCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuthSendSmsCodeLogic {
	return &AuthSendSmsCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AuthSendSmsCode 处理 POST /api/member/auth/sms/send
func (l *AuthSendSmsCodeLogic) AuthSendSmsCode(req *types.SendSmsCodeReq) (*types.SendSmsCodeResp, error) {
	if !mobileRegexp.MatchString(req.Mobile) {
		return nil, errorx.NewDefaultError("手机号格式不正确")
	}

	scene := req.Scene
	if scene == 0 {
		scene = 1
	}

	rpcResp, err := l.svcCtx.MemberAuthService.SendSmsCode(l.ctx, &umsclient.SendSmsCodeReq{
		Mobile: req.Mobile,
		Scene:  scene,
	})
	if err != nil {
		logc.Errorf(l.ctx, "发送短信验证码失败,手机号:%s,异常:%s", req.Mobile, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	expireSeconds := int32(300)
	if rpcResp != nil && rpcResp.ExpireSeconds > 0 {
		expireSeconds = rpcResp.ExpireSeconds
	}

	return &types.SendSmsCodeResp{
		Code:    0,
		Message: "验证码已发送",
		Data: types.SendSmsCodeData{
			ExpireSeconds: expireSeconds,
		},
	}, nil
}
