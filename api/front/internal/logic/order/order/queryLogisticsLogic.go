package order

import (
	"context"
	"strings"

	"github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/pkg/errorx"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/status"
)

// QueryLogisticsLogic 物流轨迹查询（Story 6-3 Task 3）
type QueryLogisticsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryLogisticsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryLogisticsLogic {
	return &QueryLogisticsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// QueryLogistics 查询物流轨迹（Story 6-3 Task 3）
// 通信拓扑: flutter-mall → front-api → OrderService.QueryOrderDetail → 提取 DeliveryData
// 降级策略: 第三方轨迹接入前，仅展示"已发货/等待揽收"状态
func (l *QueryLogisticsLogic) QueryLogistics(req *types.QueryLogisticsReq) (resp *types.QueryLogisticsResp, err error) {
	// Step 1: 归属校验 - 从 JWT 提取 memberId
	memberId, err := common.GetMemberId(l.ctx)
	if err != nil {
		return nil, err
	}

	// Step 2: 调用 OrderService.QueryOrderDetail 获取订单详情（含 DeliveryData）
	res, err := l.svcCtx.OrderService.QueryOrderDetail(l.ctx, &omsclient.QueryOrderDetailReq{
		Id:     req.OrderId,
		UserId: memberId,
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询物流失败, orderId=%d, memberId=%d, err=%s", req.OrderId, memberId, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	detail := res.Data

	// Step 3: 提取物流数据
	var deliveryCompany, deliveryNo, deliveryTime string
	if detail.DeliveryData != nil {
		deliveryCompany = detail.DeliveryData.GetDeliveryCompany()
		deliveryNo = detail.DeliveryData.GetDeliveryNo()
		deliveryTime = detail.DeliveryTime
	}

	// Step 4: 推导 currentStatus（降级展示逻辑，Story 8-6 接入第三方轨迹后动态更新）
	var currentStatus string
	var nodes []types.LogisticsNode

	if detail.DeliveryData == nil || deliveryNo == "" {
		// OMS 无物流数据（未发货/物流未填写）
		currentStatus = "暂无物流信息，包裹正在准备中"
		nodes = []types.LogisticsNode{}
	} else {
		// OMS 有物流数据但无第三方轨迹
		currentStatus = "包裹已发出，等待揽收"
		// 构建已发货节点（第三方轨迹接入前仅展示此节点）
		deliveryNodeTime := formatLogisticsTime(deliveryTime)
		nodes = []types.LogisticsNode{
			{
				Status:      "current",
				Description: "包裹已发出，等待揽收",
				Time:        deliveryNodeTime,
			},
		}
	}

	return &types.QueryLogisticsResp{
		Code:    0,
		Message: "操作成功",
		Data: types.LogisticsData{ 
			DeliveryCompany: deliveryCompany,
			DeliveryNo:     deliveryNo,
			CurrentStatus:  currentStatus,
			LogisticsNodes: nodes,
		},
	}, nil
}

// formatLogisticsTime 将 ISO 时间格式化为 YYYY-MM-DD HH:mm
func formatLogisticsTime(isoTime string) string {
	if isoTime == "" || isoTime == "0001-01-01T00:00:00Z" {
		return ""
	}
	// 格式: 2025-03-29T10:30:00+08:00 → 2025-03-29 10:30
	parts := strings.Split(isoTime, "T")
	if len(parts) < 2 {
		return isoTime
	}
	datePart := parts[0]
	timePart := parts[1]
	hourMin := strings.Split(timePart, ":")
	if len(hourMin) >= 2 {
		return datePart + " " + hourMin[0] + ":" + hourMin[1]
	}
	return isoTime
}
