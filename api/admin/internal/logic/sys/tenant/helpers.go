package tenant

import (
	"strings"

	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"google.golang.org/grpc/status"
)

func grpcError(err error) error {
	s, _ := status.FromError(err)
	return errorx.NewDefaultError(s.Message())
}

func trimStringSlice(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		normalized := strings.TrimSpace(value)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}

	return result
}

func mapTenantData(item *sysclient.TenantData) *types.TenantData {
	if item == nil {
		return &types.TenantData{}
	}

	return &types.TenantData{
		Id:                    item.Id,
		TenantCode:            item.TenantCode,
		TenantName:            item.TenantName,
		TenantShortName:       item.TenantShortName,
		ContactName:           item.ContactName,
		ContactMobile:         item.ContactMobile,
		ContactEmail:          item.ContactEmail,
		AvailableChannels:     item.AvailableChannels,
		DataRetentionDays:     item.DataRetentionDays,
		FeatureFlags:          item.FeatureFlags,
		Status:                item.Status,
		StatusReason:          item.StatusReason,
		PrimaryAdminUserId:    item.PrimaryAdminUserId,
		PrimaryAdminUserName:  item.PrimaryAdminUserName,
		PrimaryAdminMobile:    item.PrimaryAdminMobile,
		AdminActivationStatus: item.AdminActivationStatus,
		CreatedBy:             item.CreatedBy,
		CreatedAt:             item.CreatedAt,
		UpdatedBy:             item.UpdatedBy,
		UpdatedAt:             item.UpdatedAt,
	}
}
