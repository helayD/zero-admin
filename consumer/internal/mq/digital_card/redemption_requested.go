package digital_card

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	"github.com/feihua/zero-admin/rpc/oms/client/orderservice"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/feihua/zero-admin/rpc/sms/client/cardredemptionorderservice"
	"github.com/zeromicro/go-zero/core/logc"
	"gorm.io/gorm"
)

type redemptionOrderRow struct {
	ID              int64  `gorm:"column:id"`
	OrderNo         string `gorm:"column:order_no"`
	CardInstanceID  int64  `gorm:"column:card_instance_id"`
	HolderID        int64  `gorm:"column:holder_id"`
	ReceiverName    string `gorm:"column:receiver_name"`
	ReceiverPhone   string `gorm:"column:receiver_phone"`
	ReceiverAddress string `gorm:"column:receiver_address"`
	Status          string `gorm:"column:status"`
	OmsOrderID      int64  `gorm:"column:oms_order_id"`
	PlatformID      int64  `gorm:"column:platform_id"`
	TenantID        int64  `gorm:"column:tenant_id"`
	MerchantID      int64  `gorm:"column:merchant_id"`
	IsDeleted       int32  `gorm:"column:is_deleted"`
}

func (redemptionOrderRow) TableName() string {
	return "sms_card_redemption_order"
}

type cardInstanceRow struct {
	ID          int64  `gorm:"column:id"`
	TemplateID  int64  `gorm:"column:template_id"`
	AssetNo     string `gorm:"column:asset_no"`
	PlatformID  int64  `gorm:"column:platform_id"`
	TenantID    int64  `gorm:"column:tenant_id"`
	MerchantID  int64  `gorm:"column:merchant_id"`
	MemberID    int64  `gorm:"column:member_id"`
	AssetStatus string `gorm:"column:asset_status"`
	IsDeleted   int32  `gorm:"column:is_deleted"`
}

func (cardInstanceRow) TableName() string {
	return "sms_card_instance"
}

type cardTemplateRow struct {
	ID           int64  `gorm:"column:id"`
	TemplateName string `gorm:"column:template_name"`
	IsDeleted    int32  `gorm:"column:is_deleted"`
}

func (cardTemplateRow) TableName() string {
	return "sms_card_template"
}

type cardAssetLogRow struct {
	AssetInstanceID int64  `gorm:"column:asset_instance_id"`
	FromStatus      string `gorm:"column:from_status"`
	ToStatus        string `gorm:"column:to_status"`
	OperationType   string `gorm:"column:operation_type"`
	OperatorType    string `gorm:"column:operator_type"`
	TraceID         string `gorm:"column:trace_id"`
	ReasonCode      string `gorm:"column:reason_code"`
	ReasonText      string `gorm:"column:reason_text"`
	PayloadJSON     string `gorm:"column:payload_json"`
}

func (cardAssetLogRow) TableName() string {
	return "sms_card_asset_log"
}

// isRetryableError 判断错误是否可重试（网络超时、服务不可用等临时错误）
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	retryableKeywords := []string{"timeout", "deadline exceeded", "connection refused", "no such host", "temporary", "unavailable", "eof"}
	for _, kw := range retryableKeywords {
		if strings.Contains(errStr, kw) {
			return true
		}
	}
	return false
}

func recordRedemptionLog(db *gorm.DB, ctx context.Context, instanceID int64, fromStatus, toStatus, opType, reasonCode, reasonText, traceID, payload string) {
	logRow := &cardAssetLogRow{
		AssetInstanceID: instanceID,
		FromStatus:      fromStatus,
		ToStatus:        toStatus,
		OperationType:   opType,
		OperatorType:    "system",
		TraceID:         traceID,
		ReasonCode:      reasonCode,
		ReasonText:      reasonText,
		PayloadJSON:     payload,
	}
	if err := db.WithContext(ctx).Table(logRow.TableName()).Create(logRow).Error; err != nil {
		logc.Errorf(ctx, "记录提货单审计日志失败: %v", err)
	}
}

func RedemptionRequested(ctx context.Context, body []byte, db *gorm.DB, orderService cardredemptionorderservice.CardRedemptionOrderService, omsService orderservice.OrderService) error {
	if db == nil {
		return errors.New("数据库未初始化")
	}

	var event digitalcardmint.RedemptionRequestedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return err
	}
	if event.OrderID <= 0 {
		return errors.New("orderId 不能为空")
	}

	var order redemptionOrderRow
	if err := db.WithContext(ctx).
		Table(order.TableName()).
		Where("id = ? AND is_deleted = 0", event.OrderID).
		Take(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("提货单不存在: %d", event.OrderID)
		}
		return err
	}

	// C-5: 幂等保护——MQ 可能重投同一事件，若提货单已绑定 OMS 订单则直接确认消息，
	// 避免重新生成新 OrderNo 导致 OMS 端出现多张 0 元订单
	if order.OmsOrderID > 0 {
		logc.Infof(ctx, "提货单已绑定 OMS 订单，跳过重复处理, orderId=%d, omsOrderId=%d", order.ID, order.OmsOrderID)
		return nil
	}

	if order.Status != "pending" {
		// 非 pending（可能在 processing/shipped/cancelled/failed）一律忽略，由人工或专项 job 处理
		logc.Infof(ctx, "提货单状态非 pending，跳过事件, orderId=%d, status=%s", order.ID, order.Status)
		return nil
	}

	logc.Infof(ctx, "收到提货请求事件, orderId=%d, cardInstanceId=%d", event.OrderID, order.CardInstanceID)

	traceID := event.TraceID
	if traceID == "" {
		traceID = fmt.Sprintf("consumer-redemption-%d", event.OrderID)
	}

	// 更新状态为处理中
	if orderService != nil {
		_, err := orderService.UpdateRedemptionOrderStatus(ctx, &cardredemptionorderservice.UpdateRedemptionOrderStatusReq{
			OrderId: order.ID,
			Status:  "processing",
		})
		if err != nil {
			logc.Errorf(ctx, "更新提货单状态失败: %v", err)
			// 状态更新失败是可重试错误
			return err
		}
		recordRedemptionLog(db, ctx, order.CardInstanceID, "pending", "processing", "redemption_order_status_changed", "consumer_processing", "消费者更新为处理中", traceID, fmt.Sprintf(`{"orderId":%d}`, order.ID))
	}

	// 创建OMS订单
	if omsService != nil {
		// 再次校验 omsOrderId 仍为 0，防止 status 更新与创建 OMS 之间的间隙被并发请求抢先
		var latest redemptionOrderRow
		if err := db.WithContext(ctx).
			Table(latest.TableName()).
			Where("id = ? AND is_deleted = 0", order.ID).
			Take(&latest).Error; err == nil && latest.OmsOrderID > 0 {
			logc.Infof(ctx, "提货单 OMS 订单已被并发创建，跳过, orderId=%d, omsOrderId=%d", order.ID, latest.OmsOrderID)
			return nil
		}
		omsOrderID, err := createOmsOrder(ctx, db, omsService, &order)
		if err != nil {
			logc.Errorf(ctx, "创建OMS订单失败: %v", err)

			// 可重试错误：返回 error 触发 MQ 重试
			if isRetryableError(err) {
				// 回滚状态为 pending，允许下次重试
				if orderService != nil {
					_, _ = orderService.UpdateRedemptionOrderStatus(ctx, &cardredemptionorderservice.UpdateRedemptionOrderStatusReq{
						OrderId: order.ID,
						Status:  "pending",
					})
				}
				return err
			}

			// 不可重试错误：更新为失败，记录审计日志，确认消息
			if orderService != nil {
				_, _ = orderService.UpdateRedemptionOrderStatus(ctx, &cardredemptionorderservice.UpdateRedemptionOrderStatusReq{
					OrderId:    order.ID,
					Status:     "failed",
					OmsOrderId: 0,
				})
				recordRedemptionLog(db, ctx, order.CardInstanceID, "processing", "failed", "redemption_order_failed", "oms_create_failed", fmt.Sprintf("创建OMS订单失败: %v", err), traceID, fmt.Sprintf(`{"orderId":%d,"error":"%s"}`, order.ID, err.Error()))
			}
			return nil
		}

		// 更新OMS订单ID
		if orderService != nil {
			_, err = orderService.UpdateRedemptionOrderStatus(ctx, &cardredemptionorderservice.UpdateRedemptionOrderStatusReq{
				OrderId:    order.ID,
				Status:     "processing",
				OmsOrderId: omsOrderID,
			})
			if err != nil {
				logc.Errorf(ctx, "更新提货单OMS订单ID失败: %v", err)
				return err
			}
			recordRedemptionLog(db, ctx, order.CardInstanceID, "processing", "processing", "redemption_order_oms_bound", "oms_bound", "OMS订单绑定成功", traceID, fmt.Sprintf(`{"orderId":%d,"omsOrderId":%d}`, order.ID, omsOrderID))
		}

		logc.Infof(ctx, "创建OMS订单成功, orderId=%d, omsOrderId=%d", order.ID, omsOrderID)
	}

	return nil
}

func createOmsOrder(ctx context.Context, db *gorm.DB, omsService orderservice.OrderService, order *redemptionOrderRow) (int64, error) {
	// 查询卡片实例信息
	var instance cardInstanceRow
	if err := db.WithContext(ctx).
		Table(instance.TableName()).
		Where("id = ? AND is_deleted = 0", order.CardInstanceID).
		Take(&instance).Error; err != nil {
		return 0, fmt.Errorf("查询卡片实例失败: %w", err)
	}

	// 查询卡片模板信息
	var template cardTemplateRow
	if err := db.WithContext(ctx).
		Table(template.TableName()).
		Where("id = ? AND is_deleted = 0", instance.TemplateID).
		Take(&template).Error; err != nil {
		return 0, fmt.Errorf("查询卡片模板失败: %w", err)
	}

	// 构建OMS订单请求
	// C-5: 用 SMS 提货单的 OrderNo 作为 OMS 幂等键，确保 MQ 重投不会产生不同的 OMS 订单号
	orderNo := order.OrderNo

	omsReq := &omsclient.AddOrderReq{
		OrderNo:     orderNo,
		UserId:      order.HolderID,
		OrderStatus: 2, // 已支付（提货单直接创建为已支付）
		TotalAmount: 0,
		PayAmount:   0,
		SourceType:  4, // APP
		Remark:      fmt.Sprintf("提货卡提货订单,提货单号:%s", order.OrderNo),
		OrderItemData: []*omsclient.OrderItemData{
			{
				SkuName:        template.TemplateName,
				SkuQuantity:    1,
				SkuPrice:       0,
				SkuTotalAmount: 0,
				RealAmount:     0,
			},
		},
		ActivityType: "digital_card_redemption",
		ActivityId:   order.CardInstanceID,
	}

	resp, err := omsService.AddOrder(ctx, omsReq)
	if err != nil {
		return 0, fmt.Errorf("调用OMS创建订单失败: %w", err)
	}

	return resp.Id, nil
}
