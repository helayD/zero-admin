package order

import (
	"context"

	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/zeromicro/go-zero/core/logx"
)

// QueryReturnReasonListLogic 售后原因列表（Story 6-4 Task 4）
type QueryReturnReasonListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryReturnReasonListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryReturnReasonListLogic {
	return &QueryReturnReasonListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// QueryReturnReasonList 查询启用的售后原因列表（Story 6-4 Task 4）
// 通信拓扑: flutter-mall → front-api → OrderReturnReasonService.QueryOrderReturnReasonList
func (l *QueryReturnReasonListLogic) QueryReturnReasonList(req *types.QueryReturnReasonListReq) (*types.QueryReturnReasonListResp, error) {
	// 调用 OrderReturnReasonService.QueryOrderReturnReasonList 获取所有原因（包含禁用状态）
	res, err := l.svcCtx.OrderReturnReasonService.QueryOrderReturnReasonList(l.ctx, &omsclient.QueryOrderReturnReasonListReq{
		Status:   0, // 获取全部，Logic 层筛选 status==1
		PageNum:  1,
		PageSize: 100,
	})
	if err != nil {
		return nil, err
	}

	// 筛选 status==1（启用）的原因项
	var reasonList []types.ReturnReasonItem
	for _, item := range res.List {
		if item.GetStatus() == 1 {
			reasonList = append(reasonList, types.ReturnReasonItem{
				Id:   item.GetId(),
				Name: item.GetName(),
			})
		}
	}

	return &types.QueryReturnReasonListResp{
		Code:       0,
		Message:    "操作成功",
		ReasonList: reasonList,
	}, nil
}
