package customer_service_workstation

import (
	"context"

	admincommon "github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/oms/client/orderoperationlogservice"
	"github.com/zeromicro/go-zero/core/logc"
	"google.golang.org/grpc/status"

	"github.com/zeromicro/go-zero/core/logx"
)

var (
	operationTypeTextMap = map[int32]string{
		1: "创建订单",
		2: "支付订单",
		3: "发货",
		4: "确认收货",
		5: "取消订单",
		6: "退款",
	}
	operatorTypeTextMap = map[int32]string{
		1: "用户",
		2: "系统",
		3: "管理员",
	}
)

type OrderOperationLogListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrderOperationLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderOperationLogListLogic {
	return &OrderOperationLogListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// OrderOperationLogList 查询订单操作日志（客服工作台时间线）
func (l *OrderOperationLogListLogic) OrderOperationLogList(req *types.QueryOrderOperationLogListReq) (resp *types.QueryOrderOperationLogListResp, err error) {
	_, err = admincommon.ResolveQueryGovernanceScope(l.ctx, admincommon.RequestedGovernanceScope{
		ScopeType:  req.ScopeType,
		PlatformID: req.PlatformId,
		TenantID:   req.TenantId,
		MerchantID: req.MerchantId,
	})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}

	result, err := l.svcCtx.OrderOperationLogService.QueryOrderOperationLogList(l.ctx, &orderoperationlogservice.QueryOrderOperationLogListReq{
		OrderId:    req.OrderId,
		OrderNo:    req.OrderNo,
		PageNum:   req.Current,
		PageSize:  req.PageSize,
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询订单操作日志失败,参数：%+v,响应：%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	var list []*types.OrderOperationLogItem
	for _, item := range result.List {
		opTypeText := operationTypeTextMap[item.OperationType]
		if opTypeText == "" {
			opTypeText = "未知"
		}
		opTypeTextVal := operatorTypeTextMap[item.OperatorType]
		if opTypeTextVal == "" {
			opTypeTextVal = "未知"
		}
		list = append(list, &types.OrderOperationLogItem{
			Id:               item.Id,
			OrderId:         item.OrderId,
			OrderNo:         item.OrderNo,
			OperationType:   item.OperationType,
			OperationTypeText: opTypeText,
			OperatorId:      item.OperatorId,
			OperatorType:    item.OperatorType,
			OperatorTypeText: opTypeTextVal,
			OperatorNote:    item.OperatorNote,
			CreateTime:      item.CreateTime,
		})
	}

	return &types.QueryOrderOperationLogListResp{
		Code:     "000000",
		Message:  "查询成功",
		Data:     list,
		Total:    result.Total,
		Current:  req.Current,
		PageSize: req.PageSize,
		Success:  true,
	}, nil
}
