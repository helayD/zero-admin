package common

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/feihua/zero-admin/pkg/audit"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"gorm.io/gorm"
)

type userScopeRow struct {
	PlatformID int64 `gorm:"column:platform_id"`
	TenantID   int64 `gorm:"column:tenant_id"`
	MerchantID int64 `gorm:"column:merchant_id"`
}

func ResolveActorScope(ctx context.Context, db *gorm.DB, actorID int64) (pkgscope.GovernanceScope, error) {
	if actorID <= 0 {
		return pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypePlatform, pkgscope.DefaultPlatformID, 0, 0)
	}

	var row userScopeRow
	if err := db.WithContext(ctx).
		Table("sys_user").
		Select("platform_id, tenant_id, merchant_id").
		Where("id = ?", actorID).
		Take(&row).Error; err != nil {
		return pkgscope.GovernanceScope{}, err
	}

	return pkgscope.NormalizeGovernanceScope("", row.PlatformID, row.TenantID, row.MerchantID)
}

func ResolveWriteScope(ctx context.Context, db *gorm.DB, requested *smsclient.GovernanceScope, actorID int64) (pkgscope.GovernanceScope, error) {
	current, err := ResolveActorScope(ctx, db, actorID)
	if err != nil {
		return pkgscope.GovernanceScope{}, err
	}
	if requested == nil {
		return current, nil
	}

	resolved, err := NormalizeProtoScope(requested)
	if err != nil {
		return pkgscope.GovernanceScope{}, err
	}
	if err := ensureWriteScopeAllowed(current, resolved); err != nil {
		return pkgscope.GovernanceScope{}, err
	}

	return resolved, nil
}

func ensureWriteScopeAllowed(current, requested pkgscope.GovernanceScope) error {
	switch current.ScopeType {
	case pkgscope.SubjectTypePlatform:
		return nil
	case pkgscope.SubjectTypeTenant, pkgscope.SubjectTypeMerchant:
		if current.SameScope(requested) {
			return nil
		}
		return errors.New("当前主体不允许切换写入范围")
	default:
		return fmt.Errorf("不支持的当前主体范围：%s", current.ScopeType)
	}
}

func ApplyCouponScope(ctx context.Context, db *gorm.DB, couponID int64, current pkgscope.GovernanceScope) error {
	return db.WithContext(ctx).
		Table("sms_coupon").
		Where("id = ?", couponID).
		Updates(map[string]interface{}{
			"platform_id": current.PlatformID,
			"tenant_id":   current.TenantID,
			"merchant_id": current.MerchantID,
		}).Error
}

func ApplyDrawActivityScope(ctx context.Context, db *gorm.DB, activityID int64, current pkgscope.GovernanceScope) error {
	return db.WithContext(ctx).
		Table("sms_draw_activity").
		Where("id = ?", activityID).
		Updates(map[string]interface{}{
			"platform_id": current.PlatformID,
			"tenant_id":   current.TenantID,
			"merchant_id": current.MerchantID,
		}).Error
}

func EnsureCouponScope(ctx context.Context, db *gorm.DB, current pkgscope.GovernanceScope, ids []int64, action string, operatorID int64, operatorName, requestSummary string) ([]pkgscope.ResourceScopeRow, error) {
	uniqueIDs := pkgscope.UniquePositiveIDs(ids)
	if len(uniqueIDs) == 0 {
		return nil, errors.New("缺少有效资源ID")
	}

	rows, err := pkgscope.LoadResourceScopeRows(ctx, db, "sms_coupon", "id", uniqueIDs)
	if err != nil {
		return nil, err
	}
	if missing := pkgscope.MissingResourceIDs(rows, uniqueIDs); len(missing) > 0 {
		return nil, errors.New("优惠券不存在")
	}

	authorized, unauthorized := pkgscope.SplitResourceScopeRows(rows, current)
	if len(unauthorized) > 0 {
		recordDeniedSecurityEvents(ctx, db, current, action, "coupon", unauthorized, operatorID, operatorName, requestSummary)
		return nil, errors.New("当前主体无权修改所选优惠券")
	}

	return authorized, nil
}

func EnsureDrawActivityScope(ctx context.Context, db *gorm.DB, current pkgscope.GovernanceScope, ids []int64, action string, operatorID int64, operatorName, requestSummary string) ([]pkgscope.ResourceScopeRow, error) {
	uniqueIDs := pkgscope.UniquePositiveIDs(ids)
	if len(uniqueIDs) == 0 {
		return nil, errors.New("缺少有效资源ID")
	}

	rows, err := pkgscope.LoadResourceScopeRows(ctx, db, "sms_draw_activity", "id", uniqueIDs)
	if err != nil {
		return nil, err
	}
	if missing := pkgscope.MissingResourceIDs(rows, uniqueIDs); len(missing) > 0 {
		return nil, errors.New("抽卡活动不存在")
	}

	authorized, unauthorized := pkgscope.SplitResourceScopeRows(rows, current)
	if len(unauthorized) > 0 {
		recordDeniedSecurityEvents(ctx, db, current, action, "draw_activity", unauthorized, operatorID, operatorName, requestSummary)
		return nil, errors.New("当前主体无权修改所选抽卡活动")
	}

	return authorized, nil
}

func recordDeniedSecurityEvents(
	ctx context.Context,
	db *gorm.DB,
	current pkgscope.GovernanceScope,
	action string,
	resourceType string,
	rows []pkgscope.ResourceScopeRow,
	operatorID int64,
	operatorName string,
	requestSummary string,
) {
	if len(rows) == 0 {
		return
	}

	traceID := audit.NewTraceID(action, rows[0].ID)
	summary := strings.TrimSpace(requestSummary)
	if len(summary) > 500 {
		summary = summary[:500]
	}

	for _, row := range rows {
		payload, _ := audit.EncodeGovernancePayload(audit.GovernancePayload{
			TraceID:        traceID,
			Action:         action,
			ResourceType:   resourceType,
			ResourceID:     row.ID,
			ScopeType:      current.ScopeType,
			PlatformID:     current.PlatformID,
			TenantID:       current.TenantID,
			MerchantID:     current.MerchantID,
			ScopeLabel:     current.Label(),
			OperatorID:     operatorID,
			OperatorName:   operatorName,
			Result:         audit.SecurityEventResultDenied,
			RequestSummary: summary,
		})
		_ = audit.RecordSecurityEvent(ctx, db, audit.SecurityEvent{
			TraceID:        traceID,
			EventType:      audit.SecurityEventTypeGovernanceDenied,
			Action:         action,
			ResourceType:   resourceType,
			ResourceID:     row.ID,
			ScopeType:      current.ScopeType,
			PlatformID:     current.PlatformID,
			TenantID:       current.TenantID,
			MerchantID:     current.MerchantID,
			OperatorID:     operatorID,
			OperatorName:   operatorName,
			RequestSummary: summary,
			Result:         audit.SecurityEventResultDenied,
			Payload:        payload,
		})
	}
}
