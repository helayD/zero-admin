package order

import (
	"context"

	"github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/pkg/errorx"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"google.golang.org/grpc/status"

	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// ReturnApplyLogic 申请退货
// 重构说明（Story 6-5）：
//  - 使用 OrderStatusService 进行状态转换校验（必须为 OrderStatusCompleted=4 或已在售后中）
//  - 新增 AddOrderOperationLog 操作日志记录（operator_type=1 用户操作）
//  - AddOrderReturn 返回 pong，需通过 QueryOrderReturnList 二次查询获取 return_id（已验证）
/*
Author: LiuFeiHua
Date: 2025/6/20 10:59
*/
type ReturnApplyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReturnApplyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReturnApplyLogic {
	return &ReturnApplyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ReturnApply 申请退货
func (l *ReturnApplyLogic) ReturnApply(req *types.ReturnApplyReq) (resp *types.ReturnApplyResp, err error) {
	userId, err := common.GetMemberId(l.ctx)
	if err != nil {
		return nil, err
	}

	// Story 6-5 Task 6.1：归属校验
	detail, err := l.svcCtx.OrderService.QueryOrderDetail(l.ctx, &omsclient.QueryOrderDetailReq{
		Id: req.OrderId,
	})
	if err != nil {
		return nil, errorx.NewDefaultError("订单查询失败")
	}
	if detail == nil || detail.Data == nil {
		return nil, errorx.NewDefaultError("订单不存在")
	}
	if detail.Data.UserId != userId {
		return nil, errorx.NewDefaultError("无权操作此订单")
	}

	// Story 6-5 Task 6.1：状态转换校验 — 必须为已完成状态（4）或已在售后中才能申请售后
	currentStatus := int(detail.Data.OrderStatus)
	if currentStatus != OrderStatusCompleted && currentStatus != OrderStatusAfterSale {
		logc.Errorf(l.ctx, "非法售后申请操作：当前状态=%d（%s），orderId=%d", currentStatus, GetStatusText(currentStatus), req.OrderId)
		return nil, errorx.NewDefaultError("当前状态不允许申请售后")
	}

	// Story 6-5 Task 6.1：组装退货商品明细
	var items []*omsclient.OrderReturnItemData
	for _, item := range req.ReturnItemData {
		items = append(items, &omsclient.OrderReturnItemData{
			ReturnId:     item.ReturnId,
			OrderId:      item.OrderId,
			OrderItemId:  item.OrderItemId,
			SkuId:        item.SkuId,
			SkuName:      item.SkuName,
			SkuPic:       item.SkuPic,
			SkuAttrs:     item.SkuAttrs,
			Quantity:     item.Quantity,
			ProductPrice: item.ProductPrice,
			RealAmount:   item.RealAmount,
			Reason:       item.Reason,
			Remark:       item.Remark,
		})
	}

	// Task 6.1：调用 AddOrderReturn（返回 pong，需二次查询获取 return_id）
	_, err = l.svcCtx.OrderReturnService.AddOrderReturn(l.ctx, &omsclient.OrderReturnReq{
		OrderId:         req.OrderId,
		MemberId:        userId,
		Status:          req.Status,
		Type:            req.Type,
		Reason:          req.Reason,
		Description:     req.Description,
		ProofPic:        req.ProofPic,
		RefundAmount:    req.RefundAmount,
		ReturnName:      req.ReturnName,
		ReturnPhone:     req.ReturnPhone,
		CompanyAddress:  req.CompanyAddress,
		Remark:          req.Remark,
		OrderReturnItem: items,
	})
	if err != nil {
		logc.Errorf(l.ctx, "申请退货失败,参数: %+v,异常：%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	// Story 6-5 Task 6.1：二次查询获取 return_id 和 return_no（AddOrderReturn 返回 pong）
	returnList, err := l.svcCtx.OrderReturnService.QueryOrderReturnList(l.ctx, &omsclient.QueryOrderReturnListReq{
		OrderId: req.OrderId,
		MemberId: userId,
		Status:  0, // 待审核
		PageNum: 1,
		PageSize: 10,
	})
	var returnId int64
	var returnNo string
	if err == nil && returnList != nil && len(returnList.List) > 0 {
		// 取最新的售后记录
		for i := len(returnList.List) - 1; i >= 0; i-- {
			item := returnList.List[i]
			if item.OrderId == req.OrderId && item.MemberId == userId {
				returnId = item.Id
				returnNo = item.ReturnNo
				break
			}
		}
	}

	// Task 6.1 新增：写入操作日志（operator_type=1 用户操作，operation_type=6 退款）
	_, _ = l.svcCtx.OrderOperationLogService.AddOrderOperationLog(l.ctx, &omsclient.AddOrderOperationLogReq{
		OrderId:       req.OrderId,
		OperatorType:  OperatorTypeUser, // 1=用户操作
		OperationType: OpApplyAfterSale, // 复用业务 Op，映射到 OMS 退款=6
		OperatorNote:  "用户申请售后",
	})

	return &types.ReturnApplyResp{
		Code:     0,
		Message:  "操作成功",
		ReturnId: returnId,
		ReturnNo: returnNo,
	}, nil
}
