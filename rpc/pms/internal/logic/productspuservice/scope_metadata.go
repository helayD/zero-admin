package productspuservicelogic

import (
	"context"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	logiccommon "github.com/feihua/zero-admin/rpc/pms/internal/logic/common"
	"gorm.io/gorm"
)

type productScopeMetadata struct {
	ScopeType  string
	PlatformID int64
	TenantID   int64
	MerchantID int64
}

func loadProductScopeMetadata(
	ctx context.Context,
	db *gorm.DB,
	current pkgscope.GovernanceScope,
	productIDs []int64,
) (map[int64]productScopeMetadata, error) {
	metadata := make(map[int64]productScopeMetadata)
	if len(productIDs) == 0 {
		return metadata, nil
	}

	rows, err := logiccommon.LoadProductVisibilityRows(ctx, db, current, productIDs)
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		metadata[row.ID] = productScopeMetadata{
			ScopeType:  buildProductScopeType(row.PlatformID, row.TenantID, row.MerchantID),
			PlatformID: row.PlatformID,
			TenantID:   row.TenantID,
			MerchantID: row.MerchantID,
		}
	}

	return metadata, nil
}
