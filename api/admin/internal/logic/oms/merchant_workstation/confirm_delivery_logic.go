package merchant_workstation

import (
	"context"

	"github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/zeromicro/go-zero/core/logc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ConfirmDeliveryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewConfirmDeliveryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfirmDeliveryLogic {
	return &ConfirmDeliveryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ConfirmDelivery 商户发货确认
// 步骤1：QueryOrderDetail → 获取收货人信息
// 步骤2：OrderDeliveryService.AddOrderDelivery → 写入物流公司+运单号到 OmsOrderDelivery 表（含 DeliveryCompany）
// 步骤3：OrderService.Delivery → 更新订单状态为已发货 + 写入操作日志
// ⚠️ 严禁只调用 OrderService.Delivery（会导致 OmsOrderDelivery.DeliveryCompany 字段缺失）
func (l *ConfirmDeliveryLogic) ConfirmDelivery(req *types.DeliveryReq) (resp *types.BaseResp, err error) {
	userId, err := common.GetUserId(l.ctx)
	if err != nil {
		return nil, err
	}

	writeScope, err := common.ResolveWriteGovernanceScope(l.ctx, common.RequestedGovernanceScope{
		ScopeType:  req.ScopeType,
		PlatformID: req.PlatformId,
		TenantID:   req.TenantId,
		MerchantID: req.MerchantId,
	})
	if err != nil {
		return nil, err
	}
	scope := common.OMSGovernanceScope(writeScope)

	// 步骤1：查询订单详情，获取收货人信息
	orderResp, err := l.svcCtx.OrderService.QueryOrderDetail(l.ctx, &omsclient.QueryOrderDetailReq{
		Id:    req.OrderId,
		Scope: scope,
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询订单详情失败,OrderId:%d,错误:%s", req.OrderId, err.Error())
		return nil, errorx.NewDefaultError("查询订单详情失败")
	}
	if orderResp.Data == nil {
		return nil, errorx.NewDefaultError("订单不存在")
	}
	orderData := orderResp.Data
	var receiverName, receiverPhone, receiverProvince, receiverCity, receiverDistrict, receiverAddress string
	if orderData.DeliveryData != nil {
		d := orderData.DeliveryData
		receiverName = d.ReceiverName
		receiverPhone = d.ReceiverPhone
		receiverProvince = d.ReceiverProvince
		receiverCity = d.ReceiverCity
		receiverDistrict = d.ReceiverDistrict
		receiverAddress = d.ReceiverAddress
	}

	// 步骤2：写入物流信息到 OmsOrderDelivery 表（含 DeliveryCompany）
	// OMS 层 scope 校验在 OrderService.Delivery 的 EnsureOrderScope 中执行
	_, err = l.svcCtx.OrderDeliveryService.AddOrderDelivery(l.ctx, &omsclient.AddOrderDeliveryReq{
		OrderId:          req.OrderId,
		OrderNo:          orderData.OrderNo,
		ReceiverName:     receiverName,
		ReceiverPhone:    receiverPhone,
		ReceiverProvince: receiverProvince,
		ReceiverCity:     receiverCity,
		ReceiverDistrict: receiverDistrict,
		ReceiverAddress:  receiverAddress,
		DeliveryCompany:  req.DeliveryCompany,
		DeliveryNo:       req.DeliverySn,
	})
	if err != nil {
		logc.Errorf(l.ctx, "商户发货写入物流记录失败,OrderId:%d,错误:%s", req.OrderId, err.Error())
		return nil, errorx.NewDefaultError("写入物流信息失败")
	}

	// 步骤3：更新订单状态为已发货 + 写入操作日志
	_, err = l.svcCtx.OrderService.Delivery(l.ctx, &omsclient.DeliveryReq{
		OrderId:    req.OrderId,
		DeliverySn: req.DeliverySn,
		OperatorId: userId,
		Scope:      scope,
	})
	if err != nil {
		logc.Errorf(l.ctx, "商户发货确认失败,OrderId:%d,错误:%s", req.OrderId, err.Error())
		return nil, errorx.NewDefaultError("发货确认失败")
	}

	return &types.BaseResp{
		Code:    "000000",
		Message: "发货成功",
	}, nil
}
