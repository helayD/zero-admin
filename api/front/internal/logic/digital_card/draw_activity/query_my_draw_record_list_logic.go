// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package draw_activity

import (
	"context"

	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryMyDrawRecordListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryMyDrawRecordListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryMyDrawRecordListLogic {
	return &QueryMyDrawRecordListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryMyDrawRecordListLogic) QueryMyDrawRecordList(req *types.QueryMyDrawRecordListReq) (resp *types.QueryMyDrawRecordListResp, err error) {
	memberID, err := currentMemberID(l.ctx)
	if err != nil {
		return nil, err
	}
	result, err := l.svcCtx.DrawParticipationService.QueryMemberDrawRecordList(l.ctx, &smsclient.QueryMemberDrawRecordListReq{
		ActivityId: req.ActivityId,
		MemberId:   memberID,
		PageNum:    req.PageNum,
		PageSize:   req.PageSize,
		Scope:      buildDrawScope(l.ctx),
	})
	if err != nil {
		return nil, rpcError(l.ctx, "查询我的抽卡记录", req, err)
	}
	return &types.QueryMyDrawRecordListResp{
		Code:    drawCodeSuccess,
		Message: "查询我的抽卡记录成功",
		Data: types.QueryMyDrawRecordListData{
			Total: result.Total,
			List:  mapDrawRecords(result.List),
		},
	}, nil
}
