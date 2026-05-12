// Story 3.1.1: front-api SmsLogin logic
//
// 前置校验:
//   - 手机号格式: ^1[3-9]\d{9}$
//   - 验证码格式: ^\d{6}$（AC-12 硬约束 6 位；mock 与真实 provider 统一 6 位）
//
// 业务委托给 ums-rpc.MemberAuthService.LoginByCode 完成验证码校验、查/建会员、JWT 签发。

package member

import (
	"context"
	"regexp"

	"github.com/feihua/zero-admin/pkg/errorx"
	"github.com/feihua/zero-admin/rpc/ums/umsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"google.golang.org/grpc/status"

	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// smsCodeRegexp 验证码格式：6 位纯数字（AC-12 硬约束）。
//   - mock provider 下发固定 "123456"
//   - 阿里云/腾讯云通常下发 6 位
var smsCodeRegexp = regexp.MustCompile(`^\d{6}$`)

type AuthLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAuthLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuthLoginLogic {
	return &AuthLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AuthLogin 处理 POST /api/member/auth/login
func (l *AuthLoginLogic) AuthLogin(req *types.SmsLoginReq, ip string) (*types.SmsLoginResp, error) {
	if !mobileRegexp.MatchString(req.Mobile) {
		return nil, errorx.NewDefaultError("手机号格式不正确")
	}
	if !smsCodeRegexp.MatchString(req.Code) {
		return nil, errorx.NewDefaultError("验证码格式不正确")
	}

	rpcResp, err := l.svcCtx.MemberAuthService.LoginByCode(l.ctx, &umsclient.LoginByCodeReq{
		Mobile: req.Mobile,
		Code:   req.Code,
		Source: req.Source,
		Ip:     ip,
	})
	if err != nil {
		logc.Errorf(l.ctx, "验证码登录失败,手机号:%s,异常:%s", req.Mobile, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	return &types.SmsLoginResp{
		Code:    0,
		Message: "登录成功",
		Data: types.LoginData{
			Token:     rpcResp.Token,
			TokenHead: "Bearer",
			IsNewUser: rpcResp.IsNewUser,
		},
	}, nil
}
