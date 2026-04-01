package chain_monitor

import (
	"context"

	admincommon "github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"google.golang.org/grpc/status"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryChainMonitorListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryChainMonitorListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryChainMonitorListLogic {
	return &QueryChainMonitorListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryChainMonitorListLogic) QueryChainMonitorList(req *types.QueryChainMonitorListReq) (resp *types.QueryChainMonitorListResp, err error) {
	queryScope, err := admincommon.ResolveQueryGovernanceScope(l.ctx, admincommon.RequestedGovernanceScope{
		ScopeType:  req.ScopeType,
		PlatformID: req.PlatformId,
		TenantID:   req.TenantId,
		MerchantID: req.MerchantId,
	})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}

	result, err := l.svcCtx.OrderService.QueryCompensationChainList(l.ctx, &omsclient.QueryCompensationChainListReq{
		PageNum:            int32(req.Current),
		PageSize:          int32(req.PageSize),
		ChainType:         req.ChainType,
		ConsistencyStage:  req.ConsistencyStage,
		ConsistencyResult: req.ConsistencyResult,
		ManualRequired:    req.ManualRequired,
		StartTime:         req.StartTime,
		EndTime:           req.EndTime,
		OrderNo:           req.OrderNo,
		Scope:             admincommon.OMSGovernanceScope(queryScope),
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询链路监控列表失败,参数：%+v,响应：%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	var chainData []*types.ChainMonitorItem
	for _, item := range result.List {
		chainData = append(chainData, &types.ChainMonitorItem{
			TraceId:        item.TraceId,
			PlatformId:     item.PlatformId,
			TenantId:       item.TenantId,
			MerchantId:     item.MerchantId,
			ChainType:      item.ChainType,
			ChainTypeText:  item.ChainTypeText,
			EntityId:       item.EntityId,
			EntityNo:       item.EntityNo,
			EntityType:     item.EntityType,
			Stage:          item.Stage,
			StageText:      item.StageText,
			Result:         item.Result,
			ResultText:     item.ResultText,
			RetryCount:     item.RetryCount,
			LastError:      item.LastError,
			LastExecuteAt:  item.LastExecuteAt,
			CreatedAt:      item.CreatedAt,
			ActorId:        item.ActorId,
		})
	}

	return &types.QueryChainMonitorListResp{
		Code:    "000000",
		Message: "查询链路监控列表成功",
		Data:    chainData,
		Current: req.Current,
		PageSize: req.PageSize,
		Total:   result.Total,
		Success: true,
	}, nil
}
