package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/feihua/zero-admin/pkg/time_util"
	"github.com/feihua/zero-admin/rpc/oms/client/orderservice"
	"github.com/feihua/zero-admin/rpc/oms/client/ordersettingservice"
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

// 一致性阶段常量（与 api/front/.../order_state_machine.go 保持同步）
const (
	consistencyStageNone       = 0 // 无一致性阶段
	consistencyStageCancelling = 4 // 取消回退中
	consistencyStageCancelled  = 5 // 取消回退完成
)

// 一致性结果常量
const (
	consistencyResultNone           = 0 // 无结果
	consistencyResultProcessing     = 1 // 处理中
	consistencyResultSucceeded      = 2 // 成功
	consistencyResultFailed         = 3 // 失败
	consistencyResultManualRequired = 4 // 需要人工介入
)

const (
	cancelCompensationKeyPrefix = "order:cancel:compensation:"
	cancelCompensationTTL      = 86400
)

func CancelTimeOutOrder(ctx context.Context, rds *redis.Redis, productSkuService productskuservice.ProductSkuService, orderService orderservice.OrderService, couponRecordService couponrecordservice.CouponRecordService, memberService memberinfoservice.MemberInfoService, settingService ordersettingservice.OrderSettingService) {

	setting, err := settingService.QueryDefaultSetting(ctx, &ordersettingservice.QueryDefaultSettingReq{})
	if err != nil {
		logc.Errorf(ctx, "查询订单设置失败,错误信息：%+v", err)
		return
	}

	overtime := setting.NormalOrderOvertime
	logc.Infof(ctx, "开始扫描超时订单,超时时间=%d分钟,当前时间=%s", overtime, time_util.TimeToStr(time.Now()))

	timeOutOrderList, err := orderService.QueryTimeOutOrderList(ctx, &omsclient.QueryTimeOutOrderListReq{
		Minute: overtime,
	})
	if err != nil {
		logc.Errorf(ctx, "查询超时订单列表失败,请求参数：%+v,错误信息：%+v", overtime, err)
		return
	}

	if len(timeOutOrderList.List) == 0 {
		logc.Infof(ctx, "暂无超时的订单,当前时间：%s", time_util.TimeToStr(time.Now()))
		return
	}

	for _, orderInfo := range timeOutOrderList.List {
		orderId := orderInfo.Id
		memberId := orderInfo.UserId

		idempotentKey := fmt.Sprintf("%s%d", cancelCompensationKeyPrefix, orderId)

		set, err := rds.SetnxExCtx(ctx, idempotentKey, "processing", cancelCompensationTTL)
		if err != nil {
			logc.Errorf(ctx, "Redis 幂等检查异常 orderId=%d err=%v，跳过此订单", orderId, err)
			continue
		}
		if !set {
			logc.Infof(ctx, "取消补偿幂等命中 orderId=%d，已处理过，跳过", orderId)
			continue
		}

		curRetryCount := orderInfo.RetryCount

		// 通知补偿开始
		_, _ = orderService.UpdateOrderConsistency(ctx, &omsclient.UpdateOrderConsistencyReq{
			OrderId:           orderId,
			ConsistencyStage:  consistencyStageCancelling,
			ConsistencyResult: consistencyResultProcessing,
			RetryCount:        curRetryCount,
			ActorId:           memberId,
		})

		// 执行业务补偿：CancelOrder → 释放库存 → 回退优惠券 → 返还积分
		compensationErrMsg, compensationSuccess := executeCompensation(ctx, idempotentKey, orderId, memberId,
			productSkuService, orderService, couponRecordService, memberService)

		if compensationSuccess {
			_, _ = orderService.UpdateOrderConsistency(ctx, &omsclient.UpdateOrderConsistencyReq{
				OrderId:           orderId,
				ConsistencyStage:  consistencyStageCancelled,
				ConsistencyResult: consistencyResultSucceeded,
				RetryCount:        curRetryCount,
				ActorId:           memberId,
			})
			err = rds.SetexCtx(ctx, idempotentKey, "completed", cancelCompensationTTL)
			if err != nil {
				logc.Errorf(ctx, "更新补偿状态为 completed 失败 orderId=%d err=%v", orderId, err)
			}
			logc.Infof(ctx, "取消用户：%d 未支付的订单：%d 成功", memberId, orderId)
		} else {
			newRetryCount := curRetryCount + 1
			var finalResult int32
			if newRetryCount >= 3 {
				finalResult = consistencyResultManualRequired
				logc.Errorf(ctx, "订单 %d 补偿失败超过 3 次, 请人工介入, lastError=%s", orderId, compensationErrMsg)
			} else {
				finalResult = consistencyResultFailed
				logc.Errorf(ctx, "订单 %d 补偿失败, 第 %d 次重试, lastError=%s", orderId, newRetryCount, compensationErrMsg)
			}
			_, _ = orderService.UpdateOrderConsistency(ctx, &omsclient.UpdateOrderConsistencyReq{
				OrderId:           orderId,
				ConsistencyStage:  consistencyStageCancelling,
				ConsistencyResult: finalResult,
				LastError:         compensationErrMsg,
				RetryCount:        newRetryCount,
				ActorId:           memberId,
			})
			_, _ = rds.DelCtx(ctx, idempotentKey)
		}
	}

}

// executeCompensation 执行补偿链路，返回 (错误信息, 是否成功)
func executeCompensation(ctx context.Context, idempotentKey string, orderId, memberId int64,
	productSkuService productskuservice.ProductSkuService, orderService orderservice.OrderService,
	couponRecordService couponrecordservice.CouponRecordService, memberService memberinfoservice.MemberInfoService) (string, bool) {

	// Step 1: CancelOrder
	resp, err := orderService.CancelOrder(ctx, &omsclient.CancelOrderReq{
		MemberId: memberId,
		OrderId:  orderId,
		Source:   "timeout",
	})
	if err != nil {
		return fmt.Sprintf("CancelOrder 失败: %s", err.Error()), false
	}

	couponIds := resp.CouponIds
	integration := resp.Integration
	stockLockData := resp.Data

	// Step 2: ReleaseStockLock
	var data []*pmsclient.UpdateSkuStockData
	for _, item := range stockLockData {
		data = append(data, &pmsclient.UpdateSkuStockData{
			Id:              item.ProductSkuId,
			ProductQuantity: item.ProductQuantity,
		})
	}
	_, err = productSkuService.ReleaseSkuStockLock(ctx, &pmsclient.UpdateSkuStockReq{Data: data})
	if err != nil {
		return fmt.Sprintf("释放库存失败: %s", err.Error()), false
	}

	// Step 3: 回退优惠券
	if len(couponIds) > 0 {
		_, err = couponRecordService.UpdateCouponRecord(ctx, &smsclient.UpdateCouponRecordReq{
			MemberId:  memberId,
			Status:    0,
			CouponIds: couponIds,
		})
		if err != nil {
			return fmt.Sprintf("更新优惠券使用状态失败: %s", err.Error()), false
		}
	}

	// Step 4: 返还积分
	member, err := memberService.QueryMemberInfoDetail(ctx, &umsclient.QueryMemberInfoDetailReq{MemberId: memberId})
	if err != nil || member == nil {
		return fmt.Sprintf("查询会员信息失败: %v", err), false
	}

	i := member.Points + integration
	_, err = memberService.UpdateMemberPoints(ctx, &umsclient.UpdateMemberPointsReq{MemberId: memberId, Points: i})
	if err != nil {
		return fmt.Sprintf("返还使用积分失败: %s", err.Error()), false
	}

	return "", true
}
