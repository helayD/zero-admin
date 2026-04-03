package commentservicelogic

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

type RestoreCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRestoreCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RestoreCommentLogic {
	return &RestoreCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// RestoreComment 恢复已屏蔽评价
func (l *RestoreCommentLogic) RestoreComment(in *pmsclient.RestoreCommentReq) (*pmsclient.RestoreCommentResp, error) {
	comment, err := loadScopedComment(l.ctx, l.svcCtx, in.Id, in.PlatformId, in.TenantId, in.MerchantId)
	if err != nil {
		return nil, err
	}
	auditSnapshot := buildCommentAuditSnapshot(comment)
	if comment.Hidden == 0 && comment.AuditStatus != 3 {
		return nil, errors.New("评价当前无需恢复")
	}

	if err = l.svcCtx.ProductCommentModel.RestoreComment(l.ctx, in.Id, in.AuditorId, in.OperatorName); err != nil {
		logc.Errorf(l.ctx, "恢复评价失败,ID:%s,异常:%s", in.Id, err.Error())
		return nil, errors.New("恢复评价失败")
	}
	if err = recordCommentAuditLogWithRollback(
		l.ctx,
		l.svcCtx,
		comment,
		"restore",
		comment.AuditStatus,
		1,
		in.AuditorId,
		in.OperatorName,
		"",
		func() error {
			return l.svcCtx.ProductCommentModel.RestoreAuditSnapshot(l.ctx, in.Id, auditSnapshot)
		},
	); err != nil {
		logc.Errorf(l.ctx, "记录评价恢复日志失败,ID:%s,异常:%s", in.Id, err.Error())
		return nil, err
	}

	updateCommentStatsAsync(l.svcCtx, comment.ProductId)
	return &pmsclient.RestoreCommentResp{Pong: "ok"}, nil
}
