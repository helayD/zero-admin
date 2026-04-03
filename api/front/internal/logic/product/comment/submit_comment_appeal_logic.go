package comment

import (
	"context"
	"strings"

	frontcommon "github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/pkg/errorx"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"google.golang.org/grpc/status"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubmitCommentAppealLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSubmitCommentAppealLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitCommentAppealLogic {
	return &SubmitCommentAppealLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// SubmitCommentAppeal 提交评价申诉。
func (l *SubmitCommentAppealLogic) SubmitCommentAppeal(req *types.SubmitCommentAppealReq) (*types.SubmitCommentAppealResp, error) {
	memberID, err := frontcommon.GetMemberId(l.ctx)
	if err != nil {
		return nil, err
	}
	scope, err := frontcommon.CurrentGovernanceScope(l.ctx)
	if err != nil {
		return nil, err
	}

	detail, err := l.svcCtx.CommentService.QueryCommentDetail(l.ctx, &pmsclient.QueryCommentDetailReq{
		Id:         req.Id,
		PlatformId: scope.PlatformID,
		TenantId:   scope.TenantID,
		MerchantId: scope.MerchantID,
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询评价详情失败,ID:%s,异常:%s", req.Id, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	operatorName := strings.TrimSpace(detail.MemberNickName)
	_, err = l.svcCtx.CommentService.SubmitCommentAppeal(l.ctx, &pmsclient.SubmitCommentAppealReq{
		Id:           req.Id,
		MemberId:     memberID,
		OperatorName: operatorName,
		AppealReason: req.AppealReason,
		PlatformId:   scope.PlatformID,
		TenantId:     scope.TenantID,
		MerchantId:   scope.MerchantID,
	})
	if err != nil {
		logc.Errorf(l.ctx, "提交评价申诉失败,参数:%+v,异常:%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	return &types.SubmitCommentAppealResp{
		Code:    0,
		Message: "申诉已提交",
	}, nil
}
