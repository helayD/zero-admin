package orderservicelogic

import (
	"context"

	"github.com/feihua/zero-admin/rpc/oms/gen/model"
	"github.com/feihua/zero-admin/rpc/oms/internal/svc"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

type EscalateChainLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewEscalateChainLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EscalateChainLogic {
	return &EscalateChainLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// EscalateChain 升级链路为需人工介入
func (l *EscalateChainLogic) EscalateChain(in *omsclient.EscalateChainReq) (*omsclient.EscalateChainResp, error) {
	if in.PlatformId == 0 || in.TenantId == 0 {
		return &omsclient.EscalateChainResp{Code: 400, Msg: "主体范围参数不完整"}, nil
	}
	if len(in.EscalateReason) == 0 {
		return &omsclient.EscalateChainResp{Code: 400, Msg: "升级原因不能为空"}, nil
	}

	var order model.OmsOrderMain
	err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ? AND is_deleted = 0", in.OrderId).First(&order).Error
	if err != nil {
		return &omsclient.EscalateChainResp{Code: 404, Msg: "订单不存在"}, nil
	}

	if !validateScopeMatch(order.PlatformID, order.TenantID, order.MerchantID, in.PlatformId, in.TenantId, in.MerchantId) {
		return &omsclient.EscalateChainResp{Code: 403, Msg: "无权操作：该订单不在您的管理范围内"}, nil
	}

	beforeStage := order.ConsistencyStage
	beforeResult := order.ConsistencyResult

	db := l.svcCtx.DB.WithContext(l.ctx).Model(&model.OmsOrderMain{}).
		Where("id = ?", in.OrderId).
		Updates(map[string]interface{}{
			"consistency_stage":   9,
			"consistency_result": 4,
			"manual_required":   1,
		})
	if db.Error != nil {
		logc.Errorf(l.ctx, "EscalateChain 升级链路失败, orderId=%d, err=%s", in.OrderId, db.Error.Error())
		return &omsclient.EscalateChainResp{Code: 500, Msg: "升级链路失败"}, nil
	}

	logc.Infof(l.ctx, "EscalateChain 升级链路为需人工介入, orderId=%d, operatorId=%d, reason=%s",
		in.OrderId, in.OperatorId, in.EscalateReason)

	return &omsclient.EscalateChainResp{
		Code:          0,
		Msg:           "链路已升级为需人工介入",
		ManualRequired: true,
		BeforeStage:  beforeStage,
		BeforeResult: beforeResult,
		AfterStage:   9,
		AfterResult:  4,
	}, nil
}
