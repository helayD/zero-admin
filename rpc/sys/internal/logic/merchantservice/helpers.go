package merchantservicelogic

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/sys/internal/merchantmodel"
)

func generateMerchantCode(now time.Time) string {
	return fmt.Sprintf("MER%s%04d", now.Format("20060102150405"), now.Nanosecond()%10000)
}

func normalizeStringList(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
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

func scopeActivationByBusinessStatus(status int32) string {
	switch status {
	case merchantmodel.MerchantBusinessEnabled:
		return scope.ActivationStatusActive
	case merchantmodel.MerchantBusinessDisabled:
		return scope.ActivationStatusDisabled
	case merchantmodel.MerchantBusinessArchived:
		return scope.ActivationStatusArchived
	default:
		return scope.ActivationStatusPending
	}
}

func validateReviewTransition(currentStatus, nextStatus int32) error {
	if currentStatus == nextStatus {
		return nil
	}

	switch currentStatus {
	case merchantmodel.MerchantReviewPending, merchantmodel.MerchantReviewMaterialRequired, merchantmodel.MerchantReviewRejected:
		switch nextStatus {
		case merchantmodel.MerchantReviewMaterialRequired, merchantmodel.MerchantReviewRejected, merchantmodel.MerchantReviewApproved:
			return nil
		}
	case merchantmodel.MerchantReviewApproved:
		if nextStatus == merchantmodel.MerchantReviewApproved {
			return nil
		}
	}

	return errors.New("非法的商户审核状态流转")
}

func validateBusinessTransition(reviewStatus, currentStatus, nextStatus int32) error {
	if currentStatus == nextStatus {
		return nil
	}

	switch nextStatus {
	case merchantmodel.MerchantBusinessEnabled:
		if reviewStatus != merchantmodel.MerchantReviewApproved {
			return errors.New("商户尚未审核通过，不能启用")
		}
		if currentStatus == merchantmodel.MerchantBusinessPendingActivation || currentStatus == merchantmodel.MerchantBusinessDisabled {
			return nil
		}
	case merchantmodel.MerchantBusinessDisabled:
		if currentStatus == merchantmodel.MerchantBusinessEnabled {
			return nil
		}
	case merchantmodel.MerchantBusinessArchived:
		if currentStatus != merchantmodel.MerchantBusinessArchived {
			return nil
		}
	}

	return errors.New("非法的商户经营状态流转")
}
