package orderservicelogic

import (
	"context"
	"fmt"
	"time"

	"github.com/feihua/zero-admin/rpc/oms/gen/model"
	"github.com/feihua/zero-admin/rpc/oms/internal/svc"
	omsclient "github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

type RetryCompensationChainLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRetryCompensationChainLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RetryCompensationChainLogic {
	return &RetryCompensationChainLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// RetryCompensationChain 重试补偿链路
// 通过重新发布 oms.order.cancelled.v1 事件触发完整的 MQ 消费链路
// 幂等性由消费者的 Redis idempotentKey 保证
func (l *RetryCompensationChainLogic) RetryCompensationChain(in *omsclient.RetryCompensationChainReq) (*omsclient.RetryCompensationChainResp, error) {
	// 1. 主体范围校验
	if !hasRequiredChainScope(in.PlatformId) {
		return &omsclient.RetryCompensationChainResp{Code: 400, Msg: "主体范围参数不完整"}, nil
	}

	// 2. 查询订单并校验主体范围
	var order model.OmsOrderMain
	err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ? AND is_deleted = 0", in.OrderId).First(&order).Error
	if err != nil {
		return &omsclient.RetryCompensationChainResp{Code: 404, Msg: "订单不存在"}, nil
	}

	// 主体范围校验
	if !validateScopeMatch(order.PlatformID, order.TenantID, order.MerchantID, in.PlatformId, in.TenantId, in.MerchantId) {
		return &omsclient.RetryCompensationChainResp{Code: 403, Msg: "无权操作：该订单不在您的管理范围内"}, nil
	}

	// 3. 仅补偿链路（consistency_stage=4）才可重试
	if order.ConsistencyStage != 4 {
		return &omsclient.RetryCompensationChainResp{Code: 400, Msg: "该链路状态不允许重试"}, nil
	}

	// 4. 重试上限校验（maxRetry=3）
	if order.RetryCount >= 3 {
		return &omsclient.RetryCompensationChainResp{Code: 400, Msg: "已达最大重试次数（3次），请使用升级操作"}, nil
	}

	// 5. 检查暂停状态
	if order.Paused == 1 {
		return &omsclient.RetryCompensationChainResp{Code: 400, Msg: "链路已暂停，请先回放后继续"}, nil
	}

	// 6. 更新状态为处理中，增加重试次数
	newRetryCount := order.RetryCount + 1
	traceID := fmt.Sprintf("order-comp-%d-retry-%d", in.OrderId, newRetryCount)

	result := l.svcCtx.DB.WithContext(l.ctx).Model(&model.OmsOrderMain{}).
		Where("id = ?", in.OrderId).
		Updates(map[string]interface{}{
			"consistency_stage":    4,
			"consistency_result":   1,
			"last_compensation_at": time.Now(),
			"retry_count":          newRetryCount,
		})
	if result.Error != nil {
		logc.Errorf(l.ctx, "RetryCompensationChain 更新链路状态失败, orderId=%d, err=%s", in.OrderId, result.Error.Error())
		return &omsclient.RetryCompensationChainResp{Code: 500, Msg: "更新链路状态失败"}, nil
	}

	// 7. 重新发布补偿事件（触发完整 MQ 消费链路）
	current := buildOrderGovernanceScope(&order)

	sendOrderEvent(l.ctx, l.svcCtx, "order.cancel.queue", "order.cancelled.key",
		"oms.order.cancelled.v1", "manual_retry", in.OrderId, current, in.OperatorId,
		map[string]interface{}{
			"orderNo":    order.OrderNo,
			"retryCount": newRetryCount,
			"remark":     in.Remark,
		})

	logc.Infof(l.ctx, "RetryCompensationChain 发布重试事件, orderId=%d, retryCount=%d, traceId=%s",
		in.OrderId, newRetryCount, traceID)

	return &omsclient.RetryCompensationChainResp{
		Code:          0,
		Msg:           "重试成功，链路已重新触发",
		NewRetryCount: int64(newRetryCount),
		TraceId:       traceID,
		BeforeStage:   order.ConsistencyStage,
		BeforeResult:  order.ConsistencyResult,
		AfterStage:    4,
		AfterResult:   1,
	}, nil
}

// validateScopeMatch 校验主体范围匹配
func validateScopeMatch(orderPlatformID, orderTenantID, orderMerchantID, reqPlatformID, reqTenantID, reqMerchantID int64) bool {
	if reqPlatformID != orderPlatformID {
		return false
	}
	if reqTenantID != 0 && reqTenantID != orderTenantID {
		return false
	}
	if reqMerchantID != 0 && reqMerchantID != orderMerchantID {
		return false
	}
	return true
}
