package scope

import (
	"context"
	"sort"

	"gorm.io/gorm"
)

type ResourceScopeRow struct {
	ID         int64 `gorm:"column:id"`
	PlatformID int64 `gorm:"column:platform_id"`
	TenantID   int64 `gorm:"column:tenant_id"`
	MerchantID int64 `gorm:"column:merchant_id"`
}

func UniquePositiveIDs(ids []int64) []int64 {
	if len(ids) == 0 {
		return nil
	}

	seen := make(map[int64]struct{}, len(ids))
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })

	return result
}

func LoadResourceScopeRows(ctx context.Context, db *gorm.DB, table, idColumn string, ids []int64) ([]ResourceScopeRow, error) {
	uniqueIDs := UniquePositiveIDs(ids)
	if len(uniqueIDs) == 0 {
		return nil, nil
	}

	rows := make([]ResourceScopeRow, 0, len(uniqueIDs))
	if err := db.WithContext(ctx).
		Table(table).
		Select(idColumn+" AS id, platform_id, tenant_id, merchant_id").
		Where(idColumn+" IN ?", uniqueIDs).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	return rows, nil
}

func MissingResourceIDs(rows []ResourceScopeRow, ids []int64) []int64 {
	expected := UniquePositiveIDs(ids)
	if len(expected) == 0 {
		return nil
	}

	seen := make(map[int64]struct{}, len(rows))
	for _, row := range rows {
		seen[row.ID] = struct{}{}
	}

	missing := make([]int64, 0)
	for _, id := range expected {
		if _, ok := seen[id]; ok {
			continue
		}
		missing = append(missing, id)
	}

	return missing
}

func SplitResourceScopeRows(rows []ResourceScopeRow, current GovernanceScope) (authorized []ResourceScopeRow, unauthorized []ResourceScopeRow) {
	if len(rows) == 0 {
		return nil, nil
	}

	if current.ScopeType == SubjectTypePlatform {
		return rows, nil
	}

	authorized = make([]ResourceScopeRow, 0, len(rows))
	unauthorized = make([]ResourceScopeRow, 0)
	for _, row := range rows {
		rowScope := DefaultScope(row.PlatformID, row.TenantID, row.MerchantID)
		if current.SameScope(rowScope) {
			authorized = append(authorized, row)
			continue
		}
		unauthorized = append(unauthorized, row)
	}

	return authorized, unauthorized
}

func ResourceIDs(rows []ResourceScopeRow) []int64 {
	if len(rows) == 0 {
		return nil
	}

	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}

	return ids
}
