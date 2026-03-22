package merchant

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

func mapMerchantData(item *sysclient.MerchantData) *types.MerchantData {
	if item == nil {
		return &types.MerchantData{}
	}

	return &types.MerchantData{
		Id:                 item.Id,
		TenantId:           item.TenantId,
		TenantCode:         item.TenantCode,
		TenantName:         item.TenantName,
		TenantStatus:       item.TenantStatus,
		MerchantCode:       item.MerchantCode,
		MerchantName:       item.MerchantName,
		MerchantShortName:  item.MerchantShortName,
		ContactName:        item.ContactName,
		ContactMobile:      item.ContactMobile,
		ContactEmail:       item.ContactEmail,
		AvailableChannels:  item.AvailableChannels,
		CapabilityFlags:    item.CapabilityFlags,
		ReviewStatus:       item.ReviewStatus,
		ReviewReason:       item.ReviewReason,
		ReviewedBy:         item.ReviewedBy,
		ReviewedByName:     item.ReviewedByName,
		ReviewedAt:         item.ReviewedAt,
		BusinessStatus:     item.BusinessStatus,
		StatusReason:       item.StatusReason,
		VisibleScopeHint:   item.VisibleScopeHint,
		PrimaryAdminUserId: item.PrimaryAdminUserId,
		Remark:             item.Remark,
		NextActions:        item.NextActions,
		CreatedBy:          item.CreatedBy,
		CreatedAt:          item.CreatedAt,
		UpdatedBy:          item.UpdatedBy,
		UpdatedAt:          item.UpdatedAt,
	}
}
