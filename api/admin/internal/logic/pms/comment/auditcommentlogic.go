package comment

import (
	"context"
	"errors"
	"strings"

	"github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/common/res"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/status"
)

type AuditCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAuditCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuditCommentLogic {
	return &AuditCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AuditCommentLogic) AuditComment(req *types.AuditCommentReq) (resp *types.BaseResp, err error) {
	if req.Id == "" {
		return nil, errors.New("评价ID不能为空")
	}
	if req.AuditStatus != 1 && req.AuditStatus != 2 && req.AuditStatus != 3 {
		return nil, errors.New("审核状态非法")
	}
	if req.AuditStatus == 2 && strings.TrimSpace(req.AuditRemark) == "" {
		return nil, errors.New("审核拒绝时必须填写原因")
	}

	userID, err := common.GetUserId(l.ctx)
	if err != nil {
		return nil, err
	}
	userName, err := common.GetUserName(l.ctx)
	if err != nil {
		return nil, err
	}
	currentScope, err := common.CurrentGovernanceScope(l.ctx)
	if err != nil {
		return nil, err
	}

	hidden := int32(0)
	if req.AuditStatus == 3 {
		hidden = 1
	}

	_, err = l.svcCtx.CommentService.UpdateComment(l.ctx, &pmsclient.UpdateCommentReq{
		Id:          req.Id,
		AuditStatus: req.AuditStatus,
		AuditRemark: req.AuditRemark,
		Hidden:      hidden,
		AuditorId:   userID,
		UpdateBy:    userName,
		PlatformId:  currentScope.PlatformID,
		TenantId:    currentScope.TenantID,
		MerchantId:  currentScope.MerchantID,
	})
	if err != nil {
		logc.Errorf(l.ctx, "审核评价失败,参数:%+v,异常:%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	return res.Success()
}
