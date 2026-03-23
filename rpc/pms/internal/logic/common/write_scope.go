package common

import (
	"context"
	"errors"
	"strings"

	"github.com/feihua/zero-admin/pkg/audit"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
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

func ResolveWriteScope(ctx context.Context, db *gorm.DB, requested *pmsclient.GovernanceScope, actorID int64) (pkgscope.GovernanceScope, error) {
	if requested != nil {
		return NormalizeProtoScope(requested)
	}

	return ResolveActorScope(ctx, db, actorID)
}

func ApplyProductScope(ctx context.Context, db *gorm.DB, productID int64, current pkgscope.GovernanceScope) error {
	scopeValues := map[string]interface{}{
		"platform_id": current.PlatformID,
		"tenant_id":   current.TenantID,
		"merchant_id": current.MerchantID,
	}

	tx := db.WithContext(ctx)
	if err := tx.Table("pms_product_spu").Where("id = ?", productID).Updates(scopeValues).Error; err != nil {
		return err
	}

	for _, update := range []struct {
		table  string
		column string
	}{
		{table: "pms_product_sku", column: "spu_id"},
		{table: "pms_product_attribute_value", column: "spu_id"},
		{table: "pms_member_price", column: "product_id"},
		{table: "pms_product_ladder", column: "product_id"},
		{table: "pms_product_full_reduction", column: "product_id"},
	} {
		if err := tx.Table(update.table).Where(update.column+" = ?", productID).Updates(scopeValues).Error; err != nil {
			return err
		}
	}

	return nil
}

func ApplySkuScope(ctx context.Context, db *gorm.DB, skuID int64, current pkgscope.GovernanceScope) error {
	return ApplySkuScopeBatch(ctx, db, []int64{skuID}, current)
}

func ApplySkuScopeBatch(ctx context.Context, db *gorm.DB, skuIDs []int64, current pkgscope.GovernanceScope) error {
	uniqueIDs := pkgscope.UniquePositiveIDs(skuIDs)
	if len(uniqueIDs) == 0 {
		return nil
	}

	return db.WithContext(ctx).
		Table("pms_product_sku").
		Where("id IN ?", uniqueIDs).
		Updates(map[string]interface{}{
			"platform_id": current.PlatformID,
			"tenant_id":   current.TenantID,
			"merchant_id": current.MerchantID,
		}).Error
}

func EnsureProductScope(ctx context.Context, db *gorm.DB, current pkgscope.GovernanceScope, ids []int64, action string, operatorID int64, operatorName, requestSummary string) ([]pkgscope.ResourceScopeRow, error) {
	return ensureWriteScope(ctx, db, current, "pms_product_spu", "id", ids, action, "product_spu", operatorID, operatorName, requestSummary, "当前主体无权修改所选商品SPU", "商品SPU不存在")
}

func EnsureSkuScope(ctx context.Context, db *gorm.DB, current pkgscope.GovernanceScope, ids []int64, action string, operatorID int64, operatorName, requestSummary string) ([]pkgscope.ResourceScopeRow, error) {
	return ensureWriteScope(ctx, db, current, "pms_product_sku", "id", ids, action, "product_sku", operatorID, operatorName, requestSummary, "当前主体无权修改所选商品SKU", "商品SKU不存在")
}

func EnsureSpuScopeForSku(ctx context.Context, db *gorm.DB, current pkgscope.GovernanceScope, spuID int64, action string, operatorID int64, operatorName, requestSummary string) error {
	_, err := EnsureProductScope(ctx, db, current, []int64{spuID}, action, operatorID, operatorName, requestSummary)
	if err != nil {
		if strings.Contains(err.Error(), "商品SPU不存在") {
			return err
		}
		return errors.New("当前主体无权在该商品SPU下维护SKU")
	}

	return nil
}

func ensureWriteScope(
	ctx context.Context,
	db *gorm.DB,
	current pkgscope.GovernanceScope,
	table string,
	idColumn string,
	ids []int64,
	action string,
	resourceType string,
	operatorID int64,
	operatorName string,
	requestSummary string,
	unauthorizedMessage string,
	notFoundMessage string,
) ([]pkgscope.ResourceScopeRow, error) {
	uniqueIDs := pkgscope.UniquePositiveIDs(ids)
	if len(uniqueIDs) == 0 {
		return nil, errors.New("缺少有效资源ID")
	}

	rows, err := pkgscope.LoadResourceScopeRows(ctx, db, table, idColumn, uniqueIDs)
	if err != nil {
		return nil, err
	}

	if missing := pkgscope.MissingResourceIDs(rows, uniqueIDs); len(missing) > 0 {
		return nil, errors.New(notFoundMessage)
	}

	authorized, unauthorized := pkgscope.SplitResourceScopeRows(rows, current)
	if len(unauthorized) > 0 {
		recordDeniedSecurityEvents(ctx, db, current, action, resourceType, unauthorized, operatorID, operatorName, requestSummary)
		return nil, errors.New(unauthorizedMessage)
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
