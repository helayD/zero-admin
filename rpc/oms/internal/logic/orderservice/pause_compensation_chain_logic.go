package orderservicelogic

import (
	"context"
	"time"

	"github.com/feihua/zero-admin/rpc/oms/gen/model"
	"github.com/feihua/zero-admin/rpc/oms/internal/svc"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

type PauseCompensationChainLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPauseCompensationChainLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PauseCompensationChainLogic {
	return &PauseCompensationChainLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// PauseCompensationChain 暂停补偿链路
func (l *PauseCompensationChainLogic) PauseCompensationChain(in *omsclient.PauseCompensationChainReq) (*omsclient.PauseCompensationChainResp, error) {
	if !hasRequiredChainScope(in.PlatformId) {
		return &omsclient.PauseCompensationChainResp{Code: 400, Msg: "主体范围参数不完整"}, nil
	}
	if len(in.PauseReason) == 0 {
		return &omsclient.PauseCompensationChainResp{Code: 400, Msg: "暂停原因不能为空"}, nil
	}

	var order model.OmsOrderMain
	err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ? AND is_deleted = 0", in.OrderId).First(&order).Error
	if err != nil {
		return &omsclient.PauseCompensationChainResp{Code: 404, Msg: "订单不存在"}, nil
	}

	if !validateScopeMatch(order.PlatformID, order.TenantID, order.MerchantID, in.PlatformId, in.TenantId, in.MerchantId) {
		return &omsclient.PauseCompensationChainResp{Code: 403, Msg: "无权操作：该订单不在您的管理范围内"}, nil
	}

	if order.Paused == 1 {
		return &omsclient.PauseCompensationChainResp{Code: 400, Msg: "链路已处于暂停状态"}, nil
	}

	beforeStage := order.ConsistencyStage
	beforeResult := order.ConsistencyResult

	result := l.svcCtx.DB.WithContext(l.ctx).Model(&model.OmsOrderMain{}).
		Where("id = ?", in.OrderId).
		Updates(map[string]interface{}{
			"paused":            1,
			"paused_at":         time.Now(),
			"pause_reason":      in.PauseReason,
			"pause_operator_id": in.OperatorId,
		})
	if result.Error != nil {
		logc.Errorf(l.ctx, "PauseCompensationChain 暂停链路失败, orderId=%d, err=%s", in.OrderId, err.Error())
		return &omsclient.PauseCompensationChainResp{Code: 500, Msg: "暂停链路失败"}, nil
	}

	logc.Infof(l.ctx, "PauseCompensationChain 暂停链路成功, orderId=%d, operatorId=%d, reason=%s",
		in.OrderId, in.OperatorId, in.PauseReason)

	return &omsclient.PauseCompensationChainResp{
		Code:         0,
		Msg:          "链路已暂停",
		Paused:       true,
		BeforeStage:  beforeStage,
		BeforeResult: beforeResult,
		AfterStage:   beforeStage,
		AfterResult:  beforeResult,
	}, nil
}
