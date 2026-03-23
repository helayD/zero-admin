package common

import (
	"context"
	"errors"
	"strings"

	"github.com/feihua/zero-admin/pkg/audit"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"gorm.io/gorm"
)

func EnsureBrandScope(ctx context.Context, db *gorm.DB, current pkgscope.GovernanceScope, ids []int64, action string, operatorID int64, operatorName, requestSummary string) ([]pkgscope.ResourceScopeRow, error) {
	return ensureCatalogWriteScope(ctx, db, current, "pms_product_brand", ids, action, "product_brand", operatorID, operatorName, requestSummary, "当前主体无权修改所选商品品牌", "商品品牌不存在")
}

func EnsureCategoryScope(ctx context.Context, db *gorm.DB, current pkgscope.GovernanceScope, ids []int64, action string, operatorID int64, operatorName, requestSummary string) ([]pkgscope.ResourceScopeRow, error) {
	return ensureCatalogWriteScope(ctx, db, current, "pms_product_category", ids, action, "product_category", operatorID, operatorName, requestSummary, "当前主体无权修改所选商品分类", "商品分类不存在")
}

func ensureCatalogWriteScope(ctx context.Context, db *gorm.DB, current pkgscope.GovernanceScope, table string, ids []int64, action string, resourceType string, operatorID int64, operatorName string, requestSummary string, unauthorizedMessage string, notFoundMessage string) ([]pkgscope.ResourceScopeRow, error) {
	uniqueIDs := pkgscope.UniquePositiveIDs(ids)
	if len(uniqueIDs) == 0 {
		return nil, errors.New("缺少有效资源ID")
	}

	rows, err := pkgscope.LoadResourceScopeRows(ctx, db, table, "id", uniqueIDs)
	if err != nil {
		return nil, err
	}
	if missing := pkgscope.MissingResourceIDs(rows, uniqueIDs); len(missing) > 0 {
		return nil, errors.New(notFoundMessage)
	}

	authorized, unauthorized := pkgscope.SplitResourceScopeRows(rows, current)
	if len(unauthorized) > 0 {
		recordCatalogDeniedSecurityEvents(ctx, db, current, action, resourceType, unauthorized, operatorID, operatorName, requestSummary)
		return nil, errors.New(unauthorizedMessage)
	}
	return authorized, nil
}

func recordCatalogDeniedSecurityEvents(ctx context.Context, db *gorm.DB, current pkgscope.GovernanceScope, action string, resourceType string, rows []pkgscope.ResourceScopeRow, operatorID int64, operatorName string, requestSummary string) {
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
