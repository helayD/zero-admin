package points

import (
	"context"

	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/rpc/ums/umsclient"
)

type QueryPointsLogLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryPointsLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryPointsLogLogic {
	return &QueryPointsLogLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// QueryPointsLog 查询我的积分变动记录
func (l *QueryPointsLogLogic) QueryPointsLog(req *types.QueryMyPointsLogReq) (*types.QueryMyPointsLogResp, error) {
	memberId, err := common.GetMemberId(l.ctx)
	if err != nil {
		return nil, err
	}
	pageNum := req.PageNum
	if pageNum <= 0 {
		pageNum = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 20
	}
	result, err := l.svcCtx.MemberPointsLogService.QueryMemberPointsLogList(l.ctx, &umsclient.QueryMemberPointsLogListReq{
		MemberId:   memberId,
		ChangeType: req.ChangeType,
		SourceType: 5, // 5 = 全部来源（RPC 中 SourceType != 5 才过滤）
		PageNum:    pageNum,
		PageSize:   pageSize,
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询积分记录失败,memberId:%d,异常:%s", memberId, err.Error())
		return nil, err
	}
	items := make([]types.PointsLogItem, 0, len(result.List))
	for _, item := range result.List {
		items = append(items, types.PointsLogItem{
			Id:           item.Id,
			ChangeType:   item.ChangeType,
			ChangePoints: item.ChangePoints,
			SourceType:   item.SourceType,
			Description:  item.Description,
			CreateTime:   item.CreateTime,
		})
	}
	return &types.QueryMyPointsLogResp{
		Total: result.Total,
		List:  items,
	}, nil
}
