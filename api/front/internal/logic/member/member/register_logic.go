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

var mobileRegexp = regexp.MustCompile(`^1[3-9]\d{9}$`)

// RegisterLogic 会员注册
/*
Author: LiuFeiHua
Date: 2025/6/19 11:10
*/
type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Register 会员注册
func (l *RegisterLogic) Register(req *types.RegisterReq) (resp *types.RegisterResp, err error) {
	// 手机号格式校验
	if !mobileRegexp.MatchString(req.Mobile) {
		return nil, errorx.NewDefaultError("手机号格式不正确")
	}
	// 密码长度校验
	if len(req.Password) < 6 {
		return nil, errorx.NewDefaultError("密码长度不能少于6位")
	}
	// 两次密码一致性校验
	if req.Password != req.ConfirmPassword {
		return nil, errorx.NewDefaultError("两次密码不一致")
	}
	rpcResult, err := l.svcCtx.MemberService.Register(l.ctx, &umsclient.RegisterReq{
		Nickname: req.Nickname,
		Mobile:   req.Mobile,
		Source:   req.Source,
		Password: req.Password,
	})

	if err != nil {
		logc.Errorf(l.ctx, "会员注册失败,手机号: %s,异常：%s", req.Mobile, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	return &types.RegisterResp{
		Code:    0,
		Message: "注册成功",
		Data: types.LoginData{
			Token:     rpcResult.Token,
			TokenHead: "Bearer",
		},
	}, nil
}
