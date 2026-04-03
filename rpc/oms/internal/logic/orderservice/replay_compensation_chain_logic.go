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
	"gorm.io/gorm"
)

type ReplayCompensationChainLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReplayCompensationChainLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReplayCompensationChainLogic {
	return &ReplayCompensationChainLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ReplayCompensationChain 回放补偿链路
// 重新发布 oms.order.cancelled.v1 事件，触发完整 MQ 消费链路
func (l *ReplayCompensationChainLogic) ReplayCompensationChain(in *omsclient.ReplayCompensationChainReq) (*omsclient.ReplayCompensationChainResp, error) {
	// 1. 主体范围校验
	if !hasRequiredChainScope(in.PlatformId) {
		return &omsclient.ReplayCompensationChainResp{Code: 400, Msg: "主体范围参数不完整"}, nil
	}

	// 2. 回放原因必填
	if len(in.ReplayReason) == 0 {
		return &omsclient.ReplayCompensationChainResp{Code: 400, Msg: "回放原因不能为空"}, nil
	}

	// 3. 查询订单
	var order model.OmsOrderMain
	err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ? AND is_deleted = 0", in.OrderId).First(&order).Error
	if err != nil {
		return &omsclient.ReplayCompensationChainResp{Code: 404, Msg: "订单不存在"}, nil
	}

	// 4. 主体范围校验
	if !validateScopeMatch(order.PlatformID, order.TenantID, order.MerchantID, in.PlatformId, in.TenantId, in.MerchantId) {
		return &omsclient.ReplayCompensationChainResp{Code: 403, Msg: "无权操作：该订单不在您的管理范围内"}, nil
	}

	// 5. 仅补偿链路（stage=4/5/9）允许回放
	if order.ConsistencyStage != 4 && order.ConsistencyStage != 5 && order.ConsistencyStage != 9 {
		return &omsclient.ReplayCompensationChainResp{Code: 400, Msg: "链路状态不允许回放，仅 stage=4/5/9 可回放"}, nil
	}

	beforeStage := order.ConsistencyStage
	beforeResult := order.ConsistencyResult

	// 6. 事务：清除暂停标记 + 重置状态
	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		// 6a. 如果是暂停链路，先清除暂停标记
		if order.Paused == 1 {
			if err := tx.Model(&model.OmsOrderMain{}).
				Where("id = ?", in.OrderId).
				Updates(map[string]interface{}{
					"paused":            0,
					"paused_at":         nil,
					"pause_reason":      "",
					"pause_operator_id": 0,
				}).Error; err != nil {
				return err
			}
			logc.Infof(l.ctx, "ReplayCompensationChain 清除暂停标记, orderId=%d", in.OrderId)
		}

		// 6b. 重置重试次数，清除错误状态
		if err := tx.Model(&model.OmsOrderMain{}).
			Where("id = ?", in.OrderId).
			Updates(map[string]interface{}{
				"consistency_stage":    4,
				"consistency_result":   1,
				"retry_count":          0,
				"last_error":           "",
				"last_compensation_at": time.Now(),
				"manual_required":      0,
			}).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		logc.Errorf(l.ctx, "ReplayCompensationChain 更新链路状态失败, orderId=%d, err=%s", in.OrderId, err.Error())
		return &omsclient.ReplayCompensationChainResp{Code: 500, Msg: "更新链路状态失败"}, nil
	}

	// 7. 重新发布补偿事件
	traceID := fmt.Sprintf("order-comp-%d-replay-%d", in.OrderId, time.Now().UnixMilli())
	current := buildOrderGovernanceScope(&order)

	sendOrderEvent(l.ctx, l.svcCtx, "order.cancel.queue", "order.cancelled.key",
		"oms.order.cancelled.v1", "manual_replay", in.OrderId, current, in.OperatorId,
		map[string]interface{}{
			"orderNo":      order.OrderNo,
			"replayReason": in.ReplayReason,
		})

	logc.Infof(l.ctx, "ReplayCompensationChain 回放链路, orderId=%d, replayReason=%s, traceId=%s",
		in.OrderId, in.ReplayReason, traceID)

	return &omsclient.ReplayCompensationChainResp{
		Code:         0,
		Msg:          "回放成功，链路已重新触发",
		TraceId:      traceID,
		BeforeStage:  beforeStage,
		BeforeResult: beforeResult,
		AfterStage:   4,
		AfterResult:  1,
	}, nil
}
