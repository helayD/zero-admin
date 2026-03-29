package order

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/pkg/errorx"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/status"
)

// ApplyAfterSalesLogic 售后申请提交（Story 6-4 Task 5）
type ApplyAfterSalesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewApplyAfterSalesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApplyAfterSalesLogic {
	return &ApplyAfterSalesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ApplyAfterSales 提交售后申请（Story 6-4 Task 5）
// 通信拓扑: flutter-mall → front-api → 归属校验 → OMS OrderReturnService
// 幂等策略: 先查询待审核售后单，若存在则直接返回已有单号
func (l *ApplyAfterSalesLogic) ApplyAfterSales(req *types.ApplyAfterSalesReq) (*types.ApplyAfterSalesResp, error) {
	// Step 1: 归属校验 - 从 JWT 提取 memberId
	memberId, err := common.GetMemberId(l.ctx)
	if err != nil {
		return nil, err
	}

	// Step 2: 调用 OrderService.QueryOrderDetail 确认订单归属和状态
	orderRes, err := l.svcCtx.OrderService.QueryOrderDetail(l.ctx, &omsclient.QueryOrderDetailReq{
		Id:     req.OrderId,
		UserId: memberId,
	})
	if err != nil {
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	detail := orderRes.Data
	// OMS order_status: 1=待支付, 2=已支付/待发货, 3=已发货, 4=已完成, 5=已取消, 6=已退款, 7=售后中
	// 允许售后的状态：4=已完成 或 7=售后中
	if detail.OrderStatus != 4 && detail.OrderStatus != 7 {
		return nil, errorx.NewDefaultError("当前订单状态不允许申请售后")
	}

	// Step 3: 幂等保护 - 查询是否已有待审核售后单
	existingList, err := l.svcCtx.OrderReturnService.QueryOrderReturnList(l.ctx, &omsclient.QueryOrderReturnListReq{
		OrderId: req.OrderId,
		MemberId: memberId,
		Status:   0, // 待审核
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询已有售后单失败, orderId=%d, memberId=%d, err=%s", req.OrderId, memberId, err.Error())
	}

	if existingList != nil && len(existingList.List) > 0 {
		// 已存在待审核售后单，幂等返回
		existing := existingList.List[0]
		return &types.ApplyAfterSalesResp{
			Code:     0,
			Message:  "已存在待审核的售后单",
			ReturnId: existing.GetId(),
			ReturnNo: existing.GetReturnNo(),
		}, nil
	}

	// Step 4: 获取退货原因名称（reasonId → reasonName）
	reasonName := ""
	if req.ReasonId > 0 {
		reasonRes, err := l.svcCtx.OrderReturnReasonService.QueryOrderReturnReasonList(l.ctx, &omsclient.QueryOrderReturnReasonListReq{
			Status:   0, // 获取全部，精确匹配 reasonId
			PageNum:  1,
			PageSize: 100,
		})
		if err != nil {
			logc.Errorf(l.ctx, "查询退货原因失败, err=%s", err.Error())
		} else {
			for _, r := range reasonRes.List {
				if r.GetId() == req.ReasonId {
					reasonName = r.GetName()
					break
				}
			}
		}
	}

	// Step 5: 调用 AddOrderReturn 提交售后申请
	// ⚠️ AddOrderReturn 返回 OrderReturnResp{ pong }，无 ReturnId / ReturnNo
	_, err = l.svcCtx.OrderReturnService.AddOrderReturn(l.ctx, &omsclient.OrderReturnReq{
		OrderId:     req.OrderId,
		MemberId:    memberId,
		Type:        req.Type,
		Reason:      reasonName, // ⚠️ proto 要求文字字符串，不能传 reasonId
		Description: req.Description,
		ProofPic:    req.ProofPics, // base64，逗号分隔
		Status:      0,              // 待审核
	})
	if err != nil {
		s, _ := status.FromError(err)
		logc.Errorf(l.ctx, "提交售后申请失败, orderId=%d, memberId=%d, err=%s", req.OrderId, memberId, err.Error())
		return nil, errorx.NewDefaultError(s.Message())
	}

	// Step 6: 再次查询获取 ReturnId 和 ReturnNo
	// ⚠️ AddOrderReturn 返回 pong，无法从返回值取 ID，必须重新查询
	newList, err := l.svcCtx.OrderReturnService.QueryOrderReturnList(l.ctx, &omsclient.QueryOrderReturnListReq{
		OrderId:  req.OrderId,
		MemberId: memberId,
		Status:   0, // 待审核
	})
	if err != nil || newList == nil || len(newList.List) == 0 {
		logc.Errorf(l.ctx, "查询新提交售后单失败, orderId=%d, memberId=%d, err=%v", req.OrderId, memberId, err)
		return nil, errors.New("售后申请已提交，但无法获取售后单号，请稍后重试")
	}

	newReturn := newList.List[0]
	return &types.ApplyAfterSalesResp{
		Code:     0,
		Message:  "操作成功",
		ReturnId: newReturn.GetId(),
		ReturnNo: newReturn.GetReturnNo(),
	}, nil
}
