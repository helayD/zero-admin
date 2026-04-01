package orderservicelogic

import (
	"context"

	"github.com/feihua/zero-admin/rpc/oms/gen/model"
	"github.com/feihua/zero-admin/rpc/oms/internal/svc"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/zeromicro/go-zero/core/logx"
)

const maxRetryCount int32 = 3

type QueryChainActionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryChainActionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryChainActionsLogic {
	return &QueryChainActionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// QueryChainActions 查询可用干预动作
func (l *QueryChainActionsLogic) QueryChainActions(in *omsclient.QueryChainActionsReq) (*omsclient.QueryChainActionsResp, error) {
	if in.PlatformId == 0 || in.TenantId == 0 {
		return &omsclient.QueryChainActionsResp{Code: 400, Msg: "主体范围参数不完整"}, nil
	}

	var order model.OmsOrderMain
	err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ? AND is_deleted = 0", in.OrderId).First(&order).Error
	if err != nil {
		return &omsclient.QueryChainActionsResp{Code: 404, Msg: "订单不存在"}, nil
	}

	if !validateScopeMatch(order.PlatformID, order.TenantID, order.MerchantID, in.PlatformId, in.TenantId, in.MerchantId) {
		return &omsclient.QueryChainActionsResp{Code: 403, Msg: "无权操作：该订单不在您的管理范围内"}, nil
	}

	var actions []string

	// 回放：任意可回放状态（stage=4/5/9）
	if order.ConsistencyStage == 4 || order.ConsistencyStage == 5 || order.ConsistencyStage == 9 {
		actions = append(actions, "replay")
	}

	// 重试：stage=4 且未达上限且未暂停
	if order.ConsistencyStage == 4 && order.RetryCount < maxRetryCount && order.Paused == 0 {
		actions = append(actions, "retry")
	}

	// 暂停：stage=4 且未暂停
	if order.ConsistencyStage == 4 && order.Paused == 0 {
		actions = append(actions, "pause")
	}

	// 升级：失败状态（result=3）或已达重试上限
	if order.ConsistencyResult == 3 || order.RetryCount >= maxRetryCount {
		actions = append(actions, "escalate")
	}

	return &omsclient.QueryChainActionsResp{
		Code:              0,
		Msg:               "查询成功",
		AvailableActions: actions,
	}, nil
}
