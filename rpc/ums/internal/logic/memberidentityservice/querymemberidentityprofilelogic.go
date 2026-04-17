package memberidentityservicelogic

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/rpc/ums/internal/svc"
	"github.com/feihua/zero-admin/rpc/ums/umsclient"

	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

type QueryMemberIdentityProfileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryMemberIdentityProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryMemberIdentityProfileLogic {
	return &QueryMemberIdentityProfileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *QueryMemberIdentityProfileLogic) QueryMemberIdentityProfile(in *umsclient.QueryMemberIdentityProfileReq) (*umsclient.QueryMemberIdentityProfileResp, error) {
	if in.MemberId <= 0 {
		return nil, errors.New("会员ID不能为空")
	}

	profile, err := LoadMemberIdentityProfile(l.ctx, l.svcCtx.DB, in.MemberId)
	if err != nil {
		logc.Errorf(l.ctx, "查询会员实名档案失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("查询会员实名档案失败")
	}

	return profile, nil
}
