package commentservicelogic

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/feihua/zero-admin/rpc/pms/gen/model"
	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

type HandleCommentAppealLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHandleCommentAppealLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleCommentAppealLogic {
	return &HandleCommentAppealLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// HandleCommentAppeal 处理评价申诉。
func (l *HandleCommentAppealLogic) HandleCommentAppeal(in *pmsclient.HandleCommentAppealReq) (*pmsclient.HandleCommentAppealResp, error) {
	if strings.TrimSpace(in.Id) == "" {
		return nil, errors.New("评价ID不能为空")
	}
	if in.AppealStatus != 2 && in.AppealStatus != 3 {
		return nil, errors.New("申诉处理状态非法")
	}
	if strings.TrimSpace(in.AppealReply) == "" {
		return nil, errors.New("申诉处理回复不能为空")
	}

	comment, err := loadScopedComment(l.ctx, l.svcCtx, in.Id, in.PlatformId, in.TenantId, in.MerchantId)
	if err != nil {
		return nil, err
	}
	if comment.AppealStatus != 1 {
		return nil, errors.New("当前评价暂无待处理申诉")
	}

	appealSnapshot := buildCommentAppealSnapshot(comment)
	auditSnapshot := buildCommentAuditSnapshot(comment)
	appealedAt := comment.AppealedAt
	handledAt := time.Now()
	appealUpdate := model.CommentAppealUpdate{
		AppealStatus:    in.AppealStatus,
		AppealReason:    comment.AppealReason,
		AppealReply:     strings.TrimSpace(in.AppealReply),
		AppealedAt:      &appealedAt,
		AppealHandledAt: &handledAt,
		UpdateBy:        in.OperatorName,
	}

	toStatus := comment.AuditStatus
	if in.AppealStatus == 2 {
		if err = l.svcCtx.ProductCommentModel.UpdateAuditStatus(l.ctx, in.Id, 1, 0, comment.AuditRemark, in.OperatorId, in.OperatorName); err != nil {
			logc.Errorf(l.ctx, "申诉通过时恢复评价失败,ID:%s,异常:%s", in.Id, err.Error())
			return nil, errors.New("处理评价申诉失败")
		}
		toStatus = 1
	}

	if err = l.svcCtx.ProductCommentModel.UpdateAppeal(l.ctx, in.Id, appealUpdate); err != nil {
		logc.Errorf(l.ctx, "更新评价申诉状态失败,ID:%s,异常:%s", in.Id, err.Error())
		if in.AppealStatus == 2 {
			if rollbackErr := l.svcCtx.ProductCommentModel.RestoreAuditSnapshot(l.ctx, in.Id, auditSnapshot); rollbackErr != nil {
				logc.Errorf(l.ctx, "申诉处理失败后回滚审核状态失败,ID:%s,异常:%s", in.Id, rollbackErr.Error())
				return nil, errors.New("处理评价申诉失败，且状态回滚失败")
			}
		}
		return nil, errors.New("处理评价申诉失败")
	}

	if err = recordCommentAuditLogWithRollback(
		l.ctx,
		l.svcCtx,
		comment,
		"appeal_handle",
		comment.AuditStatus,
		toStatus,
		in.OperatorId,
		in.OperatorName,
		appealUpdate.AppealReply,
		chainRollbackFns(
			func() error {
				return l.svcCtx.ProductCommentModel.RestoreAppealSnapshot(l.ctx, in.Id, appealSnapshot)
			},
			func() error {
				if in.AppealStatus != 2 {
					return nil
				}
				return l.svcCtx.ProductCommentModel.RestoreAuditSnapshot(l.ctx, in.Id, auditSnapshot)
			},
		),
	); err != nil {
		logc.Errorf(l.ctx, "记录评价申诉处理日志失败,ID:%s,异常:%s", in.Id, err.Error())
		return nil, err
	}

	if in.AppealStatus == 2 {
		updateCommentStatsAsync(l.svcCtx, comment.ProductId)
	}
	return &pmsclient.HandleCommentAppealResp{Pong: "ok"}, nil
}
