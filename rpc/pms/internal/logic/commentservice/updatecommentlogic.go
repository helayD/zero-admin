package commentservicelogic

import (
	"context"
	"errors"
	"strings"

	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

// UpdateCommentLogic 更新商品评价
/*
Author: LiuFeiHua
Date: 2024/6/12 16:38
Updated: 2026/04/02 - 支持审核/屏蔽批量处理与审计日志
*/
type UpdateCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCommentLogic {
	return &UpdateCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateComment 更新商品评价（支持审核/屏蔽/批量兼容接口）
func (l *UpdateCommentLogic) UpdateComment(in *pmsclient.UpdateCommentReq) (*pmsclient.UpdateCommentResp, error) {
	ids := collectCommentIDs(in.Id, in.Ids)
	if len(ids) == 0 {
		return nil, errors.New("评价ID不能为空")
	}

	var affectedCount int64
	var firstErr error
	for _, id := range ids {
		comment, err := loadScopedComment(l.ctx, l.svcCtx, id, in.PlatformId, in.TenantId, in.MerchantId)
		if err != nil {
			if len(ids) == 1 {
				return nil, err
			}
			logc.Errorf(l.ctx, "批量处理评价前校验失败,ID:%s,异常:%s", id, err.Error())
			if firstErr == nil {
				firstErr = err
			}
			continue
		}

		auditSnapshot := buildCommentAuditSnapshot(comment)
		fromStatus := comment.AuditStatus
		toStatus := fromStatus
		action := ""
		remark := in.AuditRemark

		switch {
		case in.AuditStatus > 0:
			hidden := in.Hidden
			if in.AuditStatus == 1 {
				hidden = 0
			}
			if in.AuditStatus == 3 {
				hidden = 1
			}
			if err = l.svcCtx.ProductCommentModel.UpdateAuditStatus(l.ctx, id, in.AuditStatus, hidden, in.AuditRemark, in.AuditorId, in.UpdateBy); err != nil {
				if len(ids) == 1 {
					return nil, errors.New("更新商品评价失败")
				}
				logc.Errorf(l.ctx, "批量审核评价失败,ID:%s,异常:%s", id, err.Error())
				if firstErr == nil {
					firstErr = err
				}
				continue
			}
			toStatus = in.AuditStatus
			action = auditActionByStatus(in.AuditStatus)
		default:
			if err = l.svcCtx.ProductCommentModel.UpdateStatus(l.ctx, id, in.ShowStatus, in.UpdateBy); err != nil {
				if len(ids) == 1 {
					return nil, errors.New("更新商品评价失败")
				}
				logc.Errorf(l.ctx, "批量更新评价状态失败,ID:%s,异常:%s", id, err.Error())
				if firstErr == nil {
					firstErr = err
				}
				continue
			}
			toStatus = legacyAuditStatusByShowStatus(in.ShowStatus)
			action = legacyActionByShowStatus(in.ShowStatus)
			remark = ""
		}

		if err = recordCommentAuditLogWithRollback(
			l.ctx,
			l.svcCtx,
			comment,
			action,
			fromStatus,
			toStatus,
			in.AuditorId,
			in.UpdateBy,
			remark,
			func() error {
				return l.svcCtx.ProductCommentModel.RestoreAuditSnapshot(l.ctx, id, auditSnapshot)
			},
		); err != nil {
			logc.Errorf(l.ctx, "记录评价审核日志失败,ID:%s,异常:%s", id, err.Error())
			if len(ids) == 1 {
				return nil, err
			}
			if firstErr == nil {
				firstErr = err
			}
			continue
		}

		affectedCount++
		updateCommentStatsAsync(l.svcCtx, comment.ProductId)
	}

	if affectedCount == 0 && firstErr != nil {
		return nil, firstErr
	}

	return &pmsclient.UpdateCommentResp{
		Pong:          "ok",
		AffectedCount: affectedCount,
	}, nil
}

func collectCommentIDs(id, ids string) []string {
	if strings.TrimSpace(ids) == "" {
		if strings.TrimSpace(id) == "" {
			return nil
		}
		return []string{strings.TrimSpace(id)}
	}

	parts := strings.Split(ids, ",")
	result := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}
