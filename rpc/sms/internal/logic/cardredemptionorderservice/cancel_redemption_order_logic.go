package cardredemptionorderservicelogic

import (
	"context"
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



	if order.Status != redemptionOrderStatusPending {
		return nil, fmt.Errorf("只有待提货状态的订单可以取消,当前状态:%s", order.Status)
	}

	now := time.Now()
	if err := tx.WithContext(l.ctx).
		Table(order.TableName()).
		Where("id = ?", order.ID).
		Updates(map[string]interface{}{
			"status":       redemptionOrderStatusCancelled,
			"cancel_reason": in.CancelReason,
			"updated_at":   now,
		}).Error; err != nil {
		return nil, fmt.Errorf("更新提货单状态失败: %w", err)
	}

	if err := tx.WithContext(l.ctx).
		Table(cardInstanceRow{}.TableName()).
		Where("id = ?", order.CardInstanceID).
		Updates(map[string]interface{}{
			"asset_status": cardAssetStatusClaimed,
			"update_time":  now,
		}).Error; err != nil {
		return nil, fmt.Errorf("恢复卡片状态失败: %w", err)
	}

	payload := fmt.Sprintf(`{"orderId":%d,"cancelReason":"%s"}`, order.ID, in.CancelReason)
	logRow := &cardAssetLogRow{
		AssetInstanceID:       order.CardInstanceID,
		ParticipationRecordID: 0,
		FromStatus:            cardAssetStatusPendingRedemption,
		ToStatus:              cardAssetStatusClaimed,
		OperationType:         cardAssetOperationRedemptionOrderCancelled,
		OperatorType:          "member",
		TraceID:               "",
		ReasonCode:            cardAssetReasonRedemptionCancelled,
		ReasonText:            "持有人取消提货单",
		PayloadJSON:           payload,
	}
	if err := tx.WithContext(l.ctx).Table(logRow.TableName()).Create(logRow).Error; err != nil {
		return nil, fmt.Errorf("记录资产日志失败: %w", err)
	}

	order.Status = redemptionOrderStatusCancelled
	order.CancelReason = in.CancelReason
	return &order, nil
}
