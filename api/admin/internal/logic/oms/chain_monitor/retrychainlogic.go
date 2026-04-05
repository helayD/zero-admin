package chain_monitor

import (
	"context"
	"strconv"
	"time"

	admincommon "github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/status"
)

// chainInterventionBusinessType 链路干预操作类型（对应 sys_operate_log.business_type）
const chainInterventionBusinessType int32 = 2 // 修改

func writeChainInterventionLog(ctx context.Context, svcCtx *svc.ServiceContext, operatorId int64,
	action, reason string, orderId int64, beforeStage, beforeResult, afterStage, afterResult int32) {
	if operatorId == 0 {
		return
	}

	current, _ := admincommon.CurrentGovernanceScope(ctx)
	_, _ = svcCtx.Operatelogservice.AddOperateLog(ctx, &sysclient.AddOperateLogReq{
		Title:         "链路干预-" + action,
		BusinessType:  chainInterventionBusinessType,
		Method:        "/api/oms/order/" + action + "Chain",
		RequestMethod: "POST",
		OperatorType:  1,
		OperateUrl:    "/api/oms/order/" + action + "Chain",
		Platform:      "admin",
		Status:        0,
		OperateTime:   time.Now().Format("2006-01-02 15:04:05"),
		OperateName:   strconv.FormatInt(operatorId, 10),
		DeptName:      strconv.FormatInt(current.TenantID, 10),
		OperateParam:  `{"orderId":` + strconv.FormatInt(orderId, 10) + `,"reason":"` + reason + `"}`,
		JsonResult: `{"beforeStage":` + strconv.Itoa(int(beforeStage)) +
			`,"beforeResult":` + strconv.Itoa(int(beforeResult)) +
			`,"afterStage":` + strconv.Itoa(int(afterStage)) +
			`,"afterResult":` + strconv.Itoa(int(afterResult)) + `}`,
		Extra: `{"action":"` + action + `","orderId":` + strconv.FormatInt(orderId, 10) + `}`,
	})
}

type RetryChainLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRetryChainLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RetryChainLogic {
	return &RetryChainLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RetryChainLogic) RetryChain(req *types.RetryChainReq) (*types.RetryChainResp, error) {
	writeScope, err := resolveChainWriteScope(l.ctx, admincommon.RequestedGovernanceScope{
		ScopeType:  req.ScopeType,
		PlatformID: req.PlatformId,
		TenantID:   req.TenantId,
		MerchantID: req.MerchantId,
	})
	if err != nil {
		return nil, err
	}

	operatorId, err := admincommon.GetUserId(l.ctx)
	if err != nil {
		logc.Errorf(l.ctx, "获取操作人ID失败: %s", err.Error())
		return nil, errorx.NewDefaultError("无法获取操作人身份，请重新登录")
	}

	result, err := l.svcCtx.OrderService.RetryCompensationChain(l.ctx, &omsclient.RetryCompensationChainReq{
		OrderId:    req.OrderId,
		PlatformId: writeScope.PlatformID,
		TenantId:   writeScope.TenantID,
		MerchantId: writeScope.MerchantID,
		OperatorId: operatorId,
		Remark:     req.Remark,
	})
	if err != nil {
		logc.Errorf(l.ctx, "重试链路失败, orderId=%d, err=%s", req.OrderId, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	if result.Code != 0 {
		return &types.RetryChainResp{
			Code:          strconv.FormatInt(result.Code, 10),
			Message:       result.Msg,
			NewRetryCount: result.NewRetryCount,
			TraceId:       result.TraceId,
			Success:       false,
		}, nil
	}

	// 写入审计日志
	writeChainInterventionLog(l.ctx, l.svcCtx, operatorId, "retry", req.Remark, req.OrderId,
		result.BeforeStage, result.BeforeResult, result.AfterStage, result.AfterResult)

	return &types.RetryChainResp{
		Code:          "000000",
		Message:       "链路重试成功",
		NewRetryCount: result.NewRetryCount,
		TraceId:       result.TraceId,
		Success:       true,
	}, nil
}
