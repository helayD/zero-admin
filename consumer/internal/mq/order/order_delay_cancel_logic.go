package order

import (
	"context"
	"errors"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/feihua/zero-admin/rpc/oms/client/orderservice"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/feihua/zero-admin/rpc/pms/client/productskuservice"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/feihua/zero-admin/rpc/sms/client/couponrecordservice"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/feihua/zero-admin/rpc/ums/client/memberinfoservice"
	"github.com/feihua/zero-admin/rpc/ums/umsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	cancelCompensationKeyPrefix = "order:cancel:compensation:"
	cancelCompensationTTL       = 86400
)

func OrderDelayCancel(ctx context.Context, body []byte, rds *redis.Redis, productSkuService productskuservice.ProductSkuService, orderService orderservice.OrderService, couponRecordService couponrecordservice.CouponRecordService, memberService memberinfoservice.MemberInfoService) error {
	var payload EventPayload
	if err := sonic.Unmarshal(body, &payload); err != nil {
		logc.Errorf(ctx, "OrderDelayCancel 反序列化失败: %v", err)
		return err
	}

	ctx = payload.ToContext(ctx)
	LogWithEventPayload(ctx, "OrderDelayCancel 收到延时取消事件, entityId=%d, action=%s", payload.EntityID, payload.Action)

	orderId := payload.EntityID
	memberId := payload.ActorID

	idempotentKey := fmt.Sprintf("%s%d", cancelCompensationKeyPrefix, orderId)
	set, err := rds.SetnxExCtx(ctx, idempotentKey, "processing", cancelCompensationTTL)
	if err != nil {
		logc.Errorf(ctx, "Redis 幂等检查异常 orderId=%d err=%v", orderId, err)
		return err
	}
	if !set {
		LogWithEventPayload(ctx, "取消补偿幂等命中 orderId=%d，已处理过，跳过", orderId)
		return nil
	}

	resp, err := orderService.CancelOrder(ctx, &omsclient.CancelOrderReq{
		MemberId: memberId,
		OrderId:  orderId,
		Source:   "timeout",
	})
	if err != nil {
		logc.Errorf(ctx, "CancelOrder 失败,orderId=%d,错误信息：%+v", orderId, err)
		_, _ = rds.DelCtx(ctx, idempotentKey)
		return err
	}

	couponIds := resp.CouponIds
	integration := resp.Integration
	stockLockData := resp.Data

	var data []*pmsclient.UpdateSkuStockData
	for _, item := range stockLockData {
		data = append(data, &pmsclient.UpdateSkuStockData{
			Id:              item.ProductSkuId,
			ProductQuantity: item.ProductQuantity,
		})
	}
	_, err = productSkuService.ReleaseSkuStockLock(ctx, &pmsclient.UpdateSkuStockReq{
		Data: data,
	})
	if err != nil {
		logc.Errorf(ctx, "释放库存失败,orderId=%d,错误信息：%+v", orderId, err)
		_, _ = rds.DelCtx(ctx, idempotentKey)
		return err
	}

	if len(couponIds) > 0 {
		_, err = couponRecordService.UpdateCouponRecord(ctx, &smsclient.UpdateCouponRecordReq{
			MemberId:  memberId,
			Status:    0,
			CouponIds: couponIds,
		})
		if err != nil {
			logc.Errorf(ctx, "更新优惠券使用状态失败,orderId=%d,错误信息：%+v", orderId, err)
			_, _ = rds.DelCtx(ctx, idempotentKey)
			return err
		}
	}

	member, err := memberService.QueryMemberInfoDetail(ctx, &umsclient.QueryMemberInfoDetailReq{MemberId: memberId})
	if err != nil || member == nil {
		logc.Errorf(ctx, "查询会员信息失败,orderId=%d,err=%v", orderId, err)
		_, _ = rds.DelCtx(ctx, idempotentKey)
		return fmt.Errorf("查询会员信息失败: %w", err)
	}
	i := member.Points + integration
	_, err = memberService.UpdateMemberPoints(ctx, &umsclient.UpdateMemberPointsReq{MemberId: memberId, Points: i})
	if err != nil {
		logc.Errorf(ctx, "返还使用积分失败,orderId=%d,错误信息：%+v", orderId, err)
		_, _ = rds.DelCtx(ctx, idempotentKey)
		return err
	}

	err = rds.SetexCtx(ctx, idempotentKey, "completed", cancelCompensationTTL)
	if err != nil {
		logc.Errorf(ctx, "更新补偿状态为 completed 失败 orderId=%d err=%v", orderId, err)
	}

	LogWithEventPayload(ctx, "取消用户：%d 未支付的订单：%d 成功", memberId, orderId)
	return nil
}

func OrderDelayCancelSimple(ctx context.Context, body []byte) error {
	return errors.New("OrderDelayCancelSimple not implemented yet - use OrderDelayCancel with manual ACK")
}
