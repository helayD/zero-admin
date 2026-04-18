package digital_card_chain

import (
	"context"

	admincommon "github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type QueryDigitalCardChainDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryDigitalCardChainDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryDigitalCardChainDetailLogic {
	return &QueryDigitalCardChainDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryDigitalCardChainDetailLogic) QueryDigitalCardChainDetail(req *types.QueryDigitalCardChainDetailReq) (*types.QueryDigitalCardChainDetailResp, error) {
	current, err := admincommon.CurrentGovernanceScope(l.ctx)
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}
	detail, err := l.svcCtx.CardMintAdminService.QueryTaskDetail(l.ctx, current, req.TaskId)
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}
	logs := make([]types.DigitalCardChainLogItem, 0, len(detail.Logs))
	for _, item := range detail.Logs {
		logs = append(logs, types.DigitalCardChainLogItem{
			OperationType: item.OperationType,
			OperatorType:  item.OperatorType,
			FromStatus:    item.FromStatus,
			ToStatus:      item.ToStatus,
			ReasonText:    item.ReasonText,
			TraceId:       item.TraceID,
			PayloadJson:   item.PayloadJSON,
			CreateTime:    item.CreateTime,
		})
	}
	return &types.QueryDigitalCardChainDetailResp{
		Code:    "000000",
		Message: "查询成功",
		Data: types.DigitalCardChainDetailData{
			Item:            *mapTaskItem(&detail.Item),
			RequestId:       detail.RequestID,
			ChainTxId:       detail.ChainTxID,
			LastReceiptJson: detail.LastReceiptJSON,
			Logs:            logs,
		},
		Success: true,
	}, nil
}
