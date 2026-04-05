package commentservicelogic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/feihua/zero-admin/rpc/pms/gen/model"
	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

type SubmitCommentAppealLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSubmitCommentAppealLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitCommentAppealLogic {
	return &SubmitCommentAppealLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SubmitCommentAppeal 提交评价申诉。
func (l *SubmitCommentAppealLogic) SubmitCommentAppeal(in *pmsclient.SubmitCommentAppealReq) (*pmsclient.SubmitCommentAppealResp, error) {
	if strings.TrimSpace(in.Id) == "" {
		return nil, errors.New("评价ID不能为空")
	}
	if strings.TrimSpace(in.AppealReason) == "" {
		return nil, errors.New("申诉原因不能为空")
	}

	comment, err := loadScopedComment(l.ctx, l.svcCtx, in.Id, in.PlatformId, in.TenantId, in.MerchantId)
	if err != nil {
		return nil, err
	}
	if comment.MemberId != in.MemberId {
		return nil, errors.New("无权申诉该评价")
	}
	if comment.AuditStatus == 1 && comment.Hidden == 0 {
		return nil, errors.New("当前评价无需申诉")
	}
	if comment.AppealStatus == 1 {
		return nil, errors.New("该评价已在申诉处理中")
	}

	appealSnapshot := buildCommentAppealSnapshot(comment)
	appealedAt := time.Now()
	handledAt := time.Time{}
	operatorName := resolveCommentAppealOperatorName(comment, in.OperatorName)
	update := model.CommentAppealUpdate{
		AppealStatus:    1,
		AppealReason:    strings.TrimSpace(in.AppealReason),
		AppealReply:     "",
		AppealedAt:      &appealedAt,
		AppealHandledAt: &handledAt,
		UpdateBy:        operatorName,
	}
	if err = l.svcCtx.ProductCommentModel.UpdateAppeal(l.ctx, in.Id, update); err != nil {
		logc.Errorf(l.ctx, "提交评价申诉失败,ID:%s,异常:%s", in.Id, err.Error())
		return nil, errors.New("提交评价申诉失败")
	}
	if err = recordCommentAuditLogWithRollback(
		l.ctx,
		l.svcCtx,
		comment,
		"appeal_submit",
		comment.AuditStatus,
		comment.AuditStatus,
		in.MemberId,
		operatorName,
		update.AppealReason,
		func() error {
			return l.svcCtx.ProductCommentModel.RestoreAppealSnapshot(l.ctx, in.Id, appealSnapshot)
		},
	); err != nil {
		logc.Errorf(l.ctx, "记录评价申诉日志失败,ID:%s,异常:%s", in.Id, err.Error())
		return nil, err
	}

	return &pmsclient.SubmitCommentAppealResp{Pong: "ok"}, nil
}

func resolveCommentAppealOperatorName(comment *model.ProductComment, fallback string) string {
	if strings.TrimSpace(fallback) != "" {
		return strings.TrimSpace(fallback)
	}
	if comment != nil && strings.TrimSpace(comment.MemberNickName) != "" {
		return strings.TrimSpace(comment.MemberNickName)
	}
	if comment != nil && comment.MemberId > 0 {
		return fmt.Sprintf("member-%d", comment.MemberId)
	}
	return "member"
}
