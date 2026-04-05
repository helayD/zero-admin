package commentservicelogic

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/feihua/zero-admin/pkg/time_util"
	"github.com/feihua/zero-admin/rpc/pms/gen/model"
	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/zeromicro/go-zero/core/logc"
	"gorm.io/gorm"
)

func loadScopedComment(ctx context.Context, svcCtx *svc.ServiceContext, id string, platformID, tenantID, merchantID int64) (*model.ProductComment, error) {
	item, err := svcCtx.ProductCommentModel.FindOne(ctx, id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errors.New("评价不存在")
		}
		return nil, err
	}
	if err := validateCommentScope(item, platformID, tenantID, merchantID); err != nil {
		return nil, err
	}
	return item, nil
}

func validateCommentScope(item *model.ProductComment, platformID, tenantID, merchantID int64) error {
	if platformID > 0 && item.PlatformId != platformID {
		return errors.New("无权操作该评价")
	}
	if tenantID > 0 && item.TenantId != tenantID {
		return errors.New("无权操作该评价")
	}
	if merchantID > 0 && item.MerchantId != merchantID {
		return errors.New("无权操作该评价")
	}
	return nil
}

func auditActionByStatus(auditStatus int32) string {
	switch auditStatus {
	case 1:
		return "approve"
	case 2:
		return "reject"
	case 3:
		return "hide"
	default:
		return "unknown"
	}
}

func legacyActionByShowStatus(showStatus int32) string {
	if showStatus == 1 {
		return "approve"
	}
	return "hide"
}

func legacyAuditStatusByShowStatus(showStatus int32) int32 {
	if showStatus == 1 {
		return 1
	}
	return 3
}

func legacyHiddenByShowStatus(showStatus int32) int32 {
	if showStatus == 1 {
		return 0
	}
	return 1
}

func recordCommentAuditLog(ctx context.Context, db *gorm.DB, comment *model.ProductComment, action string, fromStatus, toStatus int32, operatorID int64, operatorName, remark string) error {
	if db == nil {
		return nil
	}
	logEntry := &model.ProductCommentAuditLog{
		CommentID:    comment.ID.Hex(),
		PlatformID:   comment.PlatformId,
		TenantID:     comment.TenantId,
		MerchantID:   comment.MerchantId,
		Action:       action,
		FromStatus:   fromStatus,
		ToStatus:     toStatus,
		OperatorID:   operatorID,
		OperatorName: operatorName,
		Remark:       remark,
	}
	return db.WithContext(ctx).Create(logEntry).Error
}

func buildCommentAuditSnapshot(comment *model.ProductComment) model.CommentAuditSnapshot {
	if comment == nil {
		return model.CommentAuditSnapshot{}
	}
	return model.CommentAuditSnapshot{
		ShowStatus:  comment.ShowStatus,
		AuditStatus: comment.AuditStatus,
		Hidden:      comment.Hidden,
		AuditRemark: comment.AuditRemark,
		AuditorID:   comment.AuditorId,
		AuditorName: comment.AuditorName,
		AuditedAt:   comment.AuditedAt,
		UpdateBy:    comment.UpdateBy,
		UpdateAt:    comment.UpdateAt,
	}
}

func buildCommentAppealSnapshot(comment *model.ProductComment) model.CommentAppealSnapshot {
	if comment == nil {
		return model.CommentAppealSnapshot{}
	}
	return model.CommentAppealSnapshot{
		AppealStatus:    comment.AppealStatus,
		AppealReason:    comment.AppealReason,
		AppealReply:     comment.AppealReply,
		AppealedAt:      comment.AppealedAt,
		AppealHandledAt: comment.AppealHandledAt,
		UpdateBy:        comment.UpdateBy,
		UpdateAt:        comment.UpdateAt,
	}
}

func recordCommentAuditLogWithRollback(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	comment *model.ProductComment,
	action string,
	fromStatus, toStatus int32,
	operatorID int64,
	operatorName, remark string,
	rollback func() error,
) error {
	if err := recordCommentAuditLog(ctx, svcCtx.DB, comment, action, fromStatus, toStatus, operatorID, operatorName, remark); err != nil {
		if rollback == nil {
			return errors.New("记录评价审核日志失败")
		}
		if rollbackErr := rollback(); rollbackErr != nil {
			logc.Errorf(ctx, "记录评价审核日志失败且状态回滚失败,commentId:%s,logErr:%s,rollbackErr:%s", comment.ID.Hex(), err.Error(), rollbackErr.Error())
			return errors.New("记录评价审核日志失败，且状态回滚失败")
		}
		return errors.New("记录评价审核日志失败")
	}
	return nil
}

func chainRollbackFns(fns ...func() error) func() error {
	return func() error {
		errs := make([]string, 0)
		for _, fn := range fns {
			if fn == nil {
				continue
			}
			if err := fn(); err != nil {
				errs = append(errs, err.Error())
			}
		}
		if len(errs) > 0 {
			return errors.New(strings.Join(errs, "; "))
		}
		return nil
	}
}

func updateCommentStatsAsync(svcCtx *svc.ServiceContext, productID int64) {
	if productID <= 0 {
		return
	}
	go func() {
		updateCommentStats(context.Background(), svcCtx, productID)
	}()
}

func updateCommentStats(ctx context.Context, svcCtx *svc.ServiceContext, productID int64) {
	count, err := svcCtx.ProductCommentModel.CountByProductId(ctx, productID)
	if err != nil {
		logc.Errorf(ctx, "统计商品[%d]评价数量失败:%s", productID, err.Error())
		return
	}
	avgStar, err := svcCtx.ProductCommentModel.AvgStarByProductId(ctx, productID)
	if err != nil {
		logc.Errorf(ctx, "统计商品[%d]平均评分失败:%s", productID, err.Error())
		return
	}
	if err = svcCtx.DB.WithContext(ctx).Exec(
		"UPDATE pms_product_spu SET product_comment_count = ?, product_star = ? WHERE id = ?",
		count, avgStar, productID,
	).Error; err != nil {
		logc.Errorf(ctx, "更新商品[%d]评价统计失败:%s", productID, err.Error())
	}
}

func formatMongoTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return time_util.TimeToStr(value)
}
