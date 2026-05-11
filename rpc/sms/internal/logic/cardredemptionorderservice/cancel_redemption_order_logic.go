package cardredemptionorderservicelogic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type CancelRedemptionOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelRedemptionOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelRedemptionOrderLogic {
	return &CancelRedemptionOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CancelRedemptionOrderLogic) CancelRedemptionOrder(in *smsclient.CancelRedemptionOrderReq) (*smsclient.CancelRedemptionOrderResp, error) {
	if in.OrderId <= 0 {
		return nil, errors.New("提货单ID无效")
	}
	if in.HolderId <= 0 {
		return nil, errors.New("提货人ID无效")
	}

	var order *redemptionOrderRow
	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		var txErr error
		order, txErr = l.cancelRedemptionOrderInTx(tx, in)
		return txErr
	})
	if err != nil {
		logc.Errorf(l.ctx, "取消提货单失败: %v, orderId=%d", err, in.OrderId)
		return nil, err
	}

	return &smsclient.CancelRedemptionOrderResp{
		Order: buildRedemptionOrderData(order),
	}, nil
}

func (l *CancelRedemptionOrderLogic) cancelRedemptionOrderInTx(tx *gorm.DB, in *smsclient.CancelRedemptionOrderReq) (*redemptionOrderRow, error) {
	var order redemptionOrderRow
	if err := tx.WithContext(l.ctx).
		Table(order.TableName()).
		Where("id = ? AND holder_id = ? AND is_deleted = 0", in.OrderId, in.HolderId).
		Take(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("提货单不存在或无权操作")
		}
		return nil, err
	}

	// Review #5 H2: 关闭 fail-open——只要提货单自身带有任意作用域，
	// 请求必须精确匹配；若调用方未携带（全 0）将直接被拒绝。
	orderHasScope := order.PlatformID > 0 || order.TenantID > 0 || order.MerchantID > 0
	if orderHasScope {
		if order.PlatformID != in.PlatformId || order.TenantID != in.TenantId || order.MerchantID != in.MerchantId {
			return nil, errors.New("提货单作用域与当前用户不匹配")
		}
	}

	if order.Status != redemptionOrderStatusPending {
		return nil, fmt.Errorf("只有待提货状态的订单可以取消,当前状态:%s", order.Status)
	}

	// H-7: 已绑定 OMS 订单的提货单禁止直接取消，避免 SMS 卡片恢复 claimed 但 OMS 仍在发货
	if order.OmsOrderID > 0 {
		return nil, errors.New("提货单已对接发货系统，无法取消，请联系客服")
	}

	now := time.Now()
	// 用 CAS 控制：仅当 status 仍为 pending 且 oms_order_id 仍为 0 才能取消，
	// 防止 consumer 在状态判断与 UPDATE 之间把单子推入 processing/绑定 OMS。
	cancelUpd := tx.WithContext(l.ctx).
		Table(order.TableName()).
		Where("id = ? AND status = ? AND oms_order_id = 0 AND is_deleted = 0", order.ID, redemptionOrderStatusPending).
		Updates(map[string]interface{}{
			"status":        redemptionOrderStatusCancelled,
			"cancel_reason": in.CancelReason,
			"updated_at":    now,
		})
	if cancelUpd.Error != nil {
		return nil, fmt.Errorf("更新提货单状态失败: %w", cancelUpd.Error)
	}
	if cancelUpd.RowsAffected == 0 {
		return nil, errors.New("提货单状态已发生变化，无法取消")
	}

	// 卡片状态恢复也加 CAS，确保 instance 仍是 pending_redemption 才回滚到 claimed
	instanceUpd := tx.WithContext(l.ctx).
		Table(cardInstanceRow{}.TableName()).
		Where("id = ? AND asset_status = ? AND is_deleted = 0", order.CardInstanceID, cardAssetStatusPendingRedemption).
		Updates(map[string]interface{}{
			"asset_status": cardAssetStatusClaimed,
			"update_time":  now,
		})
	if instanceUpd.Error != nil {
		return nil, fmt.Errorf("恢复卡片状态失败: %w", instanceUpd.Error)
	}
	if instanceUpd.RowsAffected == 0 {
		return nil, errors.New("卡片状态已发生变化，请重试")
	}

	// M-3: payload 改 json.Marshal
	payloadBytes, mErr := json.Marshal(map[string]interface{}{
		"orderId":      order.ID,
		"cancelReason": in.CancelReason,
	})
	if mErr != nil {
		return nil, fmt.Errorf("序列化审计负载失败: %w", mErr)
	}
	logRow := &cardAssetLogRow{
		AssetInstanceID:       order.CardInstanceID,
		ParticipationRecordID: 0,
		FromStatus:            cardAssetStatusPendingRedemption,
		ToStatus:              cardAssetStatusClaimed,
		OperationType:         cardAssetOperationRedemptionOrderCancelled,
		OperatorType:          "member",
		TraceID:               in.TraceId,
		ReasonCode:            cardAssetReasonRedemptionCancelled,
		ReasonText:            "持有人取消提货单",
		PayloadJSON:           string(payloadBytes),
	}
	if err := tx.WithContext(l.ctx).Table(logRow.TableName()).Create(logRow).Error; err != nil {
		return nil, fmt.Errorf("记录资产日志失败: %w", err)
	}

	order.Status = redemptionOrderStatusCancelled
	order.CancelReason = in.CancelReason
	return &order, nil
}
