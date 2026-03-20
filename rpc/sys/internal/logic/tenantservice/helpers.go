package tenantservicelogic

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/sys/internal/tenantmodel"
)

func generateTenantCode(now time.Time) string {
	return fmt.Sprintf("TEN%s%04d", now.Format("20060102150405"), now.Nanosecond()%10000)
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

func activationStatusByTenantStatus(status int32) string {
	switch status {
	case tenantmodel.TenantStatusEnabled:
		return scope.ActivationStatusActive
	case tenantmodel.TenantStatusDisabled:
		return scope.ActivationStatusDisabled
	case tenantmodel.TenantStatusArchived:
		return scope.ActivationStatusArchived
	default:
		return scope.ActivationStatusPending
	}
}

func validateStatusTransition(currentStatus, nextStatus int32) error {
	if currentStatus == nextStatus {
		return nil
	}

	switch nextStatus {
	case tenantmodel.TenantStatusEnabled:
		if currentStatus == tenantmodel.TenantStatusDisabled || currentStatus == tenantmodel.TenantStatusPendingActivation {
			return nil
		}
	case tenantmodel.TenantStatusDisabled:
		if currentStatus == tenantmodel.TenantStatusEnabled || currentStatus == tenantmodel.TenantStatusPendingActivation {
			return nil
		}
	case tenantmodel.TenantStatusArchived:
		if currentStatus != tenantmodel.TenantStatusArchived {
			return nil
		}
	}

	return errors.New("非法的租户状态流转")
}
