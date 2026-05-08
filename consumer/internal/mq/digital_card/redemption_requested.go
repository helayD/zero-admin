package digital_card

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

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

	if order.Status != "pending" {
		return fmt.Errorf("提货单状态不是待处理: %s", order.Status)
	}

	logc.Infof(ctx, "收到提货请求事件, orderId=%d, cardInstanceId=%d", event.OrderID, order.CardInstanceID)

	// 更新状态为处理中
	if orderService != nil {
		_, err := orderService.UpdateRedemptionOrderStatus(ctx, &cardredemptionorderservice.UpdateRedemptionOrderStatusReq{
			OrderId: order.ID,
			Status:  "processing",
		})
		if err != nil {
			logc.Errorf(ctx, "更新提货单状态失败: %v", err)
			return err
		}
	}

	// 创建OMS订单
	if omsService != nil {
		omsOrderID, err := createOmsOrder(ctx, db, omsService, &order)
		if err != nil {
			logc.Errorf(ctx, "创建OMS订单失败: %v", err)
			// 更新状态为失败，但不返回错误，避免消息重试
			if orderService != nil {
				_, _ = orderService.UpdateRedemptionOrderStatus(ctx, &cardredemptionorderservice.UpdateRedemptionOrderStatusReq{
					OrderId: order.ID,
					Status:  "cancelled",
				})
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
			}
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
	now := time.Now()
	orderNo := fmt.Sprintf("RDO%s%d", now.Format("20060102150405"), order.ID)

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
				SkuName:     template.TemplateName,
				SkuQuantity: 1,
				SkuPrice:    0,
				SkuTotalAmount: 0,
				RealAmount:  0,
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
