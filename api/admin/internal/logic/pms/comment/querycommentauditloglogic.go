package comment

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/status"
)

type QueryCommentAuditLogLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryCommentAuditLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryCommentAuditLogLogic {
	return &QueryCommentAuditLogLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryCommentAuditLogLogic) QueryCommentAuditLog(req *types.QueryCommentAuditLogReq) (resp *types.QueryCommentAuditLogResp, err error) {
	if req.Id == "" {
		return nil, errors.New("评价ID不能为空")
	}

	currentScope, err := common.CurrentGovernanceScope(l.ctx)
	if err != nil {
		return nil, err
	}

	result, err := l.svcCtx.CommentService.QueryCommentAuditLog(l.ctx, &pmsclient.QueryCommentAuditLogReq{
		CommentId:  req.Id,
		PageNum:    int64(req.Current),
		PageSize:   int64(req.PageSize),
		PlatformId: currentScope.PlatformID,
		TenantId:   currentScope.TenantID,
		MerchantId: currentScope.MerchantID,
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询评价审核日志失败,参数:%+v,异常:%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	data := make([]*types.CommentAuditLogData, 0, len(result.List))
	for _, item := range result.List {
		data = append(data, &types.CommentAuditLogData{
			Id:           item.Id,
			CommentId:    item.CommentId,
			Action:       item.Action,
			FromStatus:   item.FromStatus,
			ToStatus:     item.ToStatus,
			OperatorId:   item.OperatorId,
			OperatorName: item.OperatorName,
			Remark:       item.Remark,
			CreatedAt:    item.CreatedAt,
		})
	}

	return &types.QueryCommentAuditLogResp{
		Code:     "000000",
		Message:  "查询成功",
		Current:  req.Current,
		Data:     data,
		PageSize: req.PageSize,
		Success:  true,
		Total:    result.Total,
	}, nil
}
