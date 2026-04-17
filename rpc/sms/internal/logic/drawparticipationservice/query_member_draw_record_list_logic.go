package drawparticipationservicelogic

import (
	"context"
	"errors"

	logiccommon "github.com/feihua/zero-admin/rpc/sms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryMemberDrawRecordListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryMemberDrawRecordListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryMemberDrawRecordListLogic {
	return &QueryMemberDrawRecordListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *QueryMemberDrawRecordListLogic) QueryMemberDrawRecordList(in *smsclient.QueryMemberDrawRecordListReq) (*smsclient.QueryMemberDrawRecordListResp, error) {
	if in.ActivityId <= 0 {
		return nil, errors.New("活动ID不能为空")
	}
	if in.MemberId <= 0 {
		return nil, errors.New("会员ID不能为空")
	}
	currentScope, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, err
	}
	if _, err = loadActivitySnapshot(l.ctx, l.svcCtx.DB, currentScope, in.ActivityId, false); err != nil {
		return nil, errors.New("抽卡活动不存在")
	}
	total, list, err := queryMemberRecordList(l.ctx, l.svcCtx.DB, in.ActivityId, in.MemberId, in.PageNum, in.PageSize)
	if err != nil {
		return nil, errors.New("查询抽卡参与记录失败")
	}
	return &smsclient.QueryMemberDrawRecordListResp{
		Total: total,
		List:  list,
	}, nil
}
