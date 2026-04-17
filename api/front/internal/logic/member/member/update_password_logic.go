package member

import (
	"context"
	"strings"

	"github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/pkg/errorx"
	"github.com/feihua/zero-admin/rpc/ums/umsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"google.golang.org/grpc/status"

	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdatePasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdatePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePasswordLogic {
	return &UpdatePasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdatePasswordLogic) UpdatePassword(req *types.UpdatePasswordReq) (resp *types.MemberResp, err error) {
	password := strings.TrimSpace(req.Password)
	if len(password) < 6 {
		return nil, errorx.NewDefaultError("密码长度不能少于6位")
	}

	memberId, err := common.GetMemberId(l.ctx)
	if err != nil {
		return nil, err
	}

	detail, err := l.svcCtx.MemberService.QueryMemberInfoDetail(l.ctx, &umsclient.QueryMemberInfoDetailReq{
		MemberId: memberId,
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询会员信息失败,参数memberId:%d,异常：%s", memberId, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	_, err = l.svcCtx.MemberService.UpdateMemberInfo(l.ctx, &umsclient.UpdateMemberInfoReq{
		Id:       detail.Id,
		Password: password,
	})
	if err != nil {
		logc.Errorf(l.ctx, "更新会员密码失败,参数memberId:%d,异常：%s", memberId, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	return &types.MemberResp{
		Code:    "000000",
		Message: "更新密码成功",
	}, nil
}
