package cardredemptionorderservicelogic

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

const (
	redemptionOrderStatusPending    = "pending"
	redemptionOrderStatusProcessing = "processing"
	redemptionOrderStatusShipped    = "shipped"
	redemptionOrderStatusDelivered  = "delivered"
	redemptionOrderStatusCancelled  = "cancelled"

	cardAssetStatusClaimed          = "claimed"
	cardAssetStatusPendingRedemption = "pending_redemption"

	cardAssetOperationRedemptionOrderCreated   = "redemption_order_created"
	cardAssetOperationRedemptionOrderCancelled = "redemption_order_cancelled"
	cardAssetReasonRedemptionRequested         = "redemption_requested"
	cardAssetReasonRedemptionCancelled         = "redemption_cancelled"
)

type redemptionOrderRow struct {
	ID              int64      `gorm:"column:id"`
	OrderNo         string     `gorm:"column:order_no"`
	CardInstanceID  int64      `gorm:"column:card_instance_id"`
	HolderID        int64      `gorm:"column:holder_id"`
	ReceiverName    string     `gorm:"column:receiver_name"`
	ReceiverPhone   string     `gorm:"column:receiver_phone"`
	ReceiverAddress string     `gorm:"column:receiver_address"`
	Status          string     `gorm:"column:status"`
	ShippedAt       *time.Time `gorm:"column:shipped_at"`
	DeliveredAt     *time.Time `gorm:"column:delivered_at"`
	CancelReason    string     `gorm:"column:cancel_reason"`
	OmsOrderID      int64      `gorm:"column:oms_order_id"`
	PlatformID      int64      `gorm:"column:platform_id"`
	TenantID        int64      `gorm:"column:tenant_id"`
	MerchantID      int64      `gorm:"column:merchant_id"`
	CreatedAt       *time.Time `gorm:"column:created_at"`
	UpdatedAt       *time.Time `gorm:"column:updated_at"`
	IsDeleted       int32      `gorm:"column:is_deleted"`
}

func (redemptionOrderRow) TableName() string {
	return "sms_card_redemption_order"
}

type cardInstanceRow struct {
	ID                int64  `gorm:"column:id"`
	PlatformID        int64  `gorm:"column:platform_id"`
	TenantID          int64  `gorm:"column:tenant_id"`
	MerchantID        int64  `gorm:"column:merchant_id"`
	MemberID          int64  `gorm:"column:member_id"`
	AssetStatus       string `gorm:"column:asset_status"`
	MintStatus        string `gorm:"column:mint_status"`
	SourceType        string `gorm:"column:source_type"`
	SourceID          int64  `gorm:"column:source_id"`
	FulfillmentRuleID int64  `gorm:"column:fulfillment_rule_id"`
	Transferable      int32  `gorm:"column:transferable"`
	TransferLimit     int32  `gorm:"column:transfer_limit"`
	ClaimCondition    string `gorm:"column:claim_condition"`
	RedemptionCondition string `gorm:"column:redemption_condition"`
	IsDeleted         int32  `gorm:"column:is_deleted"`
}

func (cardInstanceRow) TableName() string {
	return "sms_card_instance"
}

type cardAssetLogRow struct {
	ID                    int64  `gorm:"column:id"`
	AssetInstanceID       int64  `gorm:"column:asset_instance_id"`
	ParticipationRecordID int64  `gorm:"column:participation_record_id"`
	FromStatus            string `gorm:"column:from_status"`
	ToStatus              string `gorm:"column:to_status"`
	OperationType         string `gorm:"column:operation_type"`
	OperatorType          string `gorm:"column:operator_type"`
	TraceID               string `gorm:"column:trace_id"`
	ReasonCode            string `gorm:"column:reason_code"`
	ReasonText            string `gorm:"column:reason_text"`
	PayloadJSON           string `gorm:"column:payload_json"`
}

func (cardAssetLogRow) TableName() string {
	return "sms_card_asset_log"
}

type claimTokenRow struct {
	ID             int64      `gorm:"column:id"`
	Token          string     `gorm:"column:token"`
	CardInstanceID int64      `gorm:"column:card_instance_id"`
	Status         string     `gorm:"column:status"`
	ExpireAt       *time.Time `gorm:"column:expire_at"`
	ClaimedCount   int32      `gorm:"column:claimed_count"`
	MaxClaims      int32      `gorm:"column:max_claims"`
	IsDeleted      int32      `gorm:"column:is_deleted"`
}

func (claimTokenRow) TableName() string {
	return "sms_card_claim_token"
}

type CreateRedemptionOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateRedemptionOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateRedemptionOrderLogic {
	return &CreateRedemptionOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateRedemptionOrderLogic) CreateRedemptionOrder(in *smsclient.CreateRedemptionOrderReq) (*smsclient.CreateRedemptionOrderResp, error) {
	if in.CardInstanceId <= 0 {
		return nil, errors.New("卡片实例ID无效")
	}
	if in.HolderId <= 0 {
		return nil, errors.New("提货人ID无效")
	}
	if in.ReceiverName == "" {
		return nil, errors.New("收货人姓名不能为空")
	}
	if in.ReceiverPhone == "" {
		return nil, errors.New("收货人手机号不能为空")
	}
	if in.ReceiverAddress == "" {
		return nil, errors.New("收货地址不能为空")
	}

	var order *redemptionOrderRow
	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		var txErr error
		order, txErr = l.createRedemptionOrderInTx(tx, in)
		return txErr
	})
	if err != nil {
		logc.Errorf(l.ctx, "创建提货单失败: %v, cardInstanceId=%d", err, in.CardInstanceId)
		return nil, err
	}

	if l.svcCtx.RabbitMQ != nil {
		event := digitalcardmint.RedemptionRequestedEvent{
			EventName:      digitalcardmint.EventNameRedemptionRequested,
			OrderID:        order.ID,
			CardInstanceID: order.CardInstanceID,
			HolderID:       order.HolderID,
			RequestID:      in.RequestId,
			TraceID:        in.TraceId,
		}
		if err := l.publishRedemptionEvent(event); err != nil {
			logc.Errorf(l.ctx, "发布提货事件失败: %v", err)
		}
	}

	return &smsclient.CreateRedemptionOrderResp{
		Order: buildRedemptionOrderData(order),
	}, nil
}

func (l *CreateRedemptionOrderLogic) publishRedemptionEvent(event digitalcardmint.RedemptionRequestedEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return l.svcCtx.RabbitMQ.SendMessage(
		digitalcardmint.EventExchange,
		digitalcardmint.EventExchangeType,
		digitalcardmint.EventQueueRedemption,
		digitalcardmint.EventRoutingKeyRedemption,
		body,
	)
}

func (l *CreateRedemptionOrderLogic) createRedemptionOrderInTx(tx *gorm.DB, in *smsclient.CreateRedemptionOrderReq) (*redemptionOrderRow, error) {
	var instance cardInstanceRow
	if err := tx.WithContext(l.ctx).
		Table(instance.TableName()).
		Where("id = ? AND is_deleted = 0", in.CardInstanceId).
		Take(&instance).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("卡片实例不存在")
		}
		return nil, err
	}

	if instance.MemberID != in.HolderId {
		return nil, errors.New("只有卡片持有人才能提货")
	}

	if instance.AssetStatus != cardAssetStatusClaimed {
		return nil, fmt.Errorf("卡片状态不允许提货,当前状态:%s", instance.AssetStatus)
	}

	var activeTokenCount int64
	if err := tx.WithContext(l.ctx).
		Table(claimTokenRow{}.TableName()).
		Where("card_instance_id = ? AND status = ? AND is_deleted = 0", in.CardInstanceId, "active").
		Count(&activeTokenCount).Error; err != nil {
		return nil, err
	}
	if activeTokenCount > 0 {
		return nil, errors.New("该卡片存在有效的分享凭证,请先取消分享后再提货")
	}

	var existingOrderCount int64
	if err := tx.WithContext(l.ctx).
		Table(redemptionOrderRow{}.TableName()).
		Where("card_instance_id = ? AND status IN (?, ?) AND is_deleted = 0", in.CardInstanceId, redemptionOrderStatusPending, redemptionOrderStatusProcessing).
		Count(&existingOrderCount).Error; err != nil {
		return nil, err
	}
	if existingOrderCount > 0 {
		return nil, errors.New("该卡片已存在进行中的提货单")
	}

	orderNo := generateRedemptionOrderNo()
	now := time.Now()
	order := &redemptionOrderRow{
		OrderNo:         orderNo,
		CardInstanceID:  in.CardInstanceId,
		HolderID:        in.HolderId,
		ReceiverName:    in.ReceiverName,
		ReceiverPhone:   in.ReceiverPhone,
		ReceiverAddress: in.ReceiverAddress,
		Status:          redemptionOrderStatusPending,
		PlatformID:      in.PlatformId,
		TenantID:        in.TenantId,
		MerchantID:      in.MerchantId,
		CreatedAt:       &now,
	}
	if err := tx.WithContext(l.ctx).
		Table(order.TableName()).
		Create(order).Error; err != nil {
		return nil, fmt.Errorf("创建提货单失败: %w", err)
	}

	if err := tx.WithContext(l.ctx).
		Table(instance.TableName()).
		Where("id = ?", instance.ID).
		Updates(map[string]interface{}{
			"asset_status": cardAssetStatusPendingRedemption,
			"update_time":  now,
		}).Error; err != nil {
		return nil, fmt.Errorf("更新卡片状态失败: %w", err)
	}

	payload := fmt.Sprintf(`{"orderId":%d,"orderNo":"%s","holderId":%d}`, order.ID, order.OrderNo, order.HolderID)
	logRow := &cardAssetLogRow{
		AssetInstanceID:       instance.ID,
		ParticipationRecordID: 0,
		FromStatus:            cardAssetStatusClaimed,
		ToStatus:              cardAssetStatusPendingRedemption,
		OperationType:         cardAssetOperationRedemptionOrderCreated,
		OperatorType:          "member",
		TraceID:               in.TraceId,
		ReasonCode:            cardAssetReasonRedemptionRequested,
		ReasonText:            "持有人发起提货,创建提货单",
		PayloadJSON:           payload,
	}
	if err := tx.WithContext(l.ctx).Table(logRow.TableName()).Create(logRow).Error; err != nil {
		return nil, fmt.Errorf("记录资产日志失败: %w", err)
	}

	return order, nil
}

func generateRedemptionOrderNo() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("RDO%s%012d", time.Now().Format("20060102150405"), time.Now().UnixNano()%1000000000000)
	}
	return fmt.Sprintf("RDO%s%s", time.Now().Format("20060102150405"), hex.EncodeToString(buf))
}

func buildRedemptionOrderData(row *redemptionOrderRow) *smsclient.RedemptionOrderData {
	if row == nil {
		return nil
	}
	data := &smsclient.RedemptionOrderData{
		Id:              row.ID,
		OrderNo:         row.OrderNo,
		CardInstanceId:  row.CardInstanceID,
		HolderId:        row.HolderID,
		ReceiverName:    row.ReceiverName,
		ReceiverPhone:   row.ReceiverPhone,
		ReceiverAddress: row.ReceiverAddress,
		Status:          row.Status,
		OmsOrderId:      row.OmsOrderID,
		PlatformId:      row.PlatformID,
		TenantId:        row.TenantID,
		MerchantId:      row.MerchantID,
	}
	if row.ShippedAt != nil {
		data.ShippedAt = row.ShippedAt.Format("2006-01-02 15:04:05")
	}
	if row.DeliveredAt != nil {
		data.DeliveredAt = row.DeliveredAt.Format("2006-01-02 15:04:05")
	}
	if row.CreatedAt != nil {
		data.CreateTime = row.CreatedAt.Format("2006-01-02 15:04:05")
	}
	if row.UpdatedAt != nil {
		data.UpdateTime = row.UpdatedAt.Format("2006-01-02 15:04:05")
	}
	return data
}
