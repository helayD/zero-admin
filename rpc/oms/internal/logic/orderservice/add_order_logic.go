package orderservicelogic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/feihua/zero-admin/pkg/operatefunnel"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/oms/gen/model"
	"github.com/feihua/zero-admin/rpc/oms/internal/logic/common"
	"github.com/zeromicro/go-zero/core/logc"
	"gorm.io/gorm"

	"github.com/feihua/zero-admin/rpc/oms/internal/svc"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddOrderLogic {
	return &AddOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AddOrder 添加订单(app)
func (l *AddOrderLogic) AddOrder(in *omsclient.AddOrderReq) (*omsclient.AddOrderResp, error) {
	item := &model.OmsOrderMain{
		OrderNo:         in.OrderNo,                  // 订单编号
		UserID:          in.UserId,                   // 用户ID
		OrderStatus:     in.OrderStatus,              // 订单状态：1-待支付,2-已支付,3-已发货,4-已完成,5-已取消,6-已退款,7-售后中
		TotalAmount:     float64(in.TotalAmount),     // 订单总金额
		PromotionAmount: float64(in.PromotionAmount), // 促销金额
		CouponAmount:    float64(in.CouponAmount),    // 优惠券金额
		PointsAmount:    float64(in.PointsAmount),    // 积分金额
		DiscountAmount:  float64(in.DiscountAmount),  // 优惠金额
		FreightAmount:   float64(in.FreightAmount),   // 运费金额
		PayAmount:       float64(in.PayAmount),       // 实付金额
		SourceType:      in.SourceType,               // 订单来源：1-APP,2-PC,3-小程序
		UsePoints:       in.UsePoints,                // 下单时使用的积分
		Remark:          in.Remark,                   // 订单备注
	}

	var orderItems []*model.OmsOrderItem
	var cartItemIds []int64
	for _, data := range in.OrderItemData {
		item1 := &model.OmsOrderItem{
			OrderID:         item.ID,                       // 订单ID
			OrderNo:         item.OrderNo,                  // 订单编号
			OrderItemStatus: 1,                             // 订单商品状态：1-正常,2-退货申请中,3-已退货,4-已拒绝
			SkuID:           data.SkuId,                    // 商品SKU ID
			SkuName:         data.SkuName,                  // 商品名称
			SkuPic:          data.SkuPic,                   // 商品图片
			SkuPrice:        float64(data.SkuPrice),        // 商品单价
			SkuQuantity:     data.SkuQuantity,              // 商品数量
			SpecData:        data.SpecData,                 // 规格数据
			SkuTotalAmount:  float64(data.SkuTotalAmount),  // 商品总金额
			PromotionAmount: float64(data.PromotionAmount), // 促销分摊金额
			CouponAmount:    float64(data.CouponAmount),    // 优惠券分摊金额
			PointsAmount:    float64(data.PointsAmount),    // 积分分摊金额
			DiscountAmount:  float64(data.DiscountAmount),  // 优惠分摊金额
			RealAmount:      float64(data.RealAmount),      // 实付金额
		}

		orderItems = append(orderItems, item1)
		// 创建订单时 OrderItemData.Id 携带购物车项 ID；立即购买场景为 0，不清理购物车。
		if data.Id > 0 {
			cartItemIds = append(cartItemIds, data.Id)
		}

	}

	if len(orderItems) == 0 {
		return nil, errors.New("订单商品不能为空")
	}

	err := l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(item).Error; err != nil {
			logc.Errorf(l.ctx, "添加订单失败,参数:%+v,异常:%s", item, err.Error())
			return fmt.Errorf("添加订单失败")
		}

		if err := tx.Table("oms_order_main").
			Where("id = ?", item.ID).
			Updates(map[string]interface{}{
				"activity_type": operatefunnel.NormalizeActivityType(in.ActivityType),
				"activity_id":   in.ActivityId,
			}).Error; err != nil {
			logc.Errorf(l.ctx, "更新订单活动归因失败,orderId:%d,异常:%s", item.ID, err.Error())
			return fmt.Errorf("更新订单活动归因失败")
		}

		for _, orderItem := range orderItems {
			orderItem.OrderID = item.ID
		}

		if err := tx.CreateInBatches(orderItems, len(orderItems)).Error; err != nil {
			logc.Errorf(l.ctx, "添加订单商品失败,参数:%+v,异常:%s", orderItems, err.Error())
			return fmt.Errorf("添加订单商品失败")
		}

		if len(cartItemIds) > 0 {
			err := tx.Table("oms_cart_item").
				Where("id IN ? AND member_id = ?", cartItemIds, in.UserId).
				Updates(map[string]interface{}{
					"delete_status": 1,
					"update_time":   time.Now(),
				}).Error
			if err != nil {
				logc.Errorf(l.ctx, "删除购物车失败,参数:%+v,异常:%s", in, err.Error())
				return fmt.Errorf("删除购物车失败")
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	currentScope, err := common.ResolveActorScope(l.ctx, l.svcCtx.DB, in.UserId)
	if err != nil {
		logc.Errorf(l.ctx, "解析用户作用域失败,userId:%d,使用默认平台级作用域,异常:%s", in.UserId, err.Error())
		currentScope, _ = pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypePlatform, pkgscope.DefaultPlatformID, 0, 0)
	}

	sendOrderEvent(l.ctx, l.svcCtx, "order.create.queue", "order.created.key", "order.created", "", item.ID, currentScope, in.UserId, map[string]interface{}{
		"orderNo":     item.OrderNo,
		"totalAmount": item.TotalAmount,
	})

	return &omsclient.AddOrderResp{Id: item.ID}, nil
}
