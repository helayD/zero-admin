package channel_integration_template

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
)

type impactSummaryView struct {
	BindingTenantCount   int64    `json:"bindingTenantCount"`
	BindingMerchantCount int64    `json:"bindingMerchantCount"`
	BindingSampleLabels  []string `json:"bindingSampleLabels,omitempty"`
	LastStatusChangeTime string   `json:"lastStatusChangeTime,omitempty"`
	StatusReason         string   `json:"statusReason,omitempty"`
}

type impactSummaryEnvelope struct {
	ImpactSummary impactSummaryView `json:"impactSummary"`
}

func currentTemplateGovernanceScope(ctx context.Context) (pkgscope.GovernanceScope, error) {
	return common.CurrentGovernanceScope(ctx)
}

func resolveTemplateMutationScope(ctx context.Context, scopeType string, platformID, tenantID, merchantID int64) (pkgscope.GovernanceScope, error) {
	current, err := currentTemplateGovernanceScope(ctx)
	if err != nil {
		return pkgscope.GovernanceScope{}, err
	}

	resolved, err := pkgscope.NormalizeGovernanceScope(scopeType, platformID, tenantID, merchantID)
	if err != nil {
		return pkgscope.GovernanceScope{}, err
	}
	if resolved.PlatformID == 0 {
		resolved.PlatformID = current.PlatformID
	}

	switch current.ScopeType {
	case pkgscope.SubjectTypePlatform:
		return resolved, nil
	case pkgscope.SubjectTypeTenant:
		if resolved.ScopeType == pkgscope.SubjectTypePlatform {
			return pkgscope.GovernanceScope{}, errorx.NewDefaultError("当前主体不允许写入平台级模板")
		}
		if resolved.TenantID != current.TenantID {
			return pkgscope.GovernanceScope{}, errorx.NewDefaultError("当前主体不允许写入其他租户模板")
		}
		return resolved, nil
	case pkgscope.SubjectTypeMerchant:
		if resolved.ScopeType != pkgscope.SubjectTypeMerchant || resolved.MerchantID != current.MerchantID {
			return pkgscope.GovernanceScope{}, errorx.NewDefaultError("当前主体仅允许写入本商户模板")
		}
		return resolved, nil
	default:
		return pkgscope.GovernanceScope{}, errorx.NewDefaultError("当前登录上下文治理范围无效")
	}
}

func allowTemplateByScope(item *sysclient.ChannelIntegrationTemplateData, current pkgscope.GovernanceScope) bool {
	if item == nil {
		return false
	}
	if current.PlatformID > 0 && item.PlatformId > 0 && current.PlatformID != item.PlatformId {
		return false
	}
	switch current.ScopeType {
	case pkgscope.SubjectTypePlatform:
		return true
	case pkgscope.SubjectTypeTenant:
		if item.ScopeType == pkgscope.SubjectTypePlatform {
			return true
		}
		return item.TenantId == current.TenantID
	case pkgscope.SubjectTypeMerchant:
		if item.ScopeType == pkgscope.SubjectTypePlatform {
			return true
		}
		if item.ScopeType == pkgscope.SubjectTypeTenant {
			return item.TenantId == current.TenantID
		}
		return item.MerchantId == current.MerchantID
	default:
		return false
	}
}

func allowTemplateMutationByScope(item *sysclient.ChannelIntegrationTemplateData, current pkgscope.GovernanceScope) bool {
	if item == nil {
		return false
	}
	if current.PlatformID > 0 && item.PlatformId > 0 && current.PlatformID != item.PlatformId {
		return false
	}
	switch current.ScopeType {
	case pkgscope.SubjectTypePlatform:
		return true
	case pkgscope.SubjectTypeTenant:
		return item.TenantId == current.TenantID
	case pkgscope.SubjectTypeMerchant:
		return item.ScopeType == pkgscope.SubjectTypeMerchant && item.MerchantId == current.MerchantID
	default:
		return false
	}
}

func matchTemplateExtraFilters(item *sysclient.ChannelIntegrationTemplateData, tenantID, merchantID int64) bool {
	if item == nil {
		return false
	}
	if tenantID > 0 && item.TenantId != tenantID {
		return false
	}
	if merchantID > 0 && item.MerchantId != merchantID {
		return false
	}
	return true
}

func parseImpactSummary(raw string) impactSummaryView {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return impactSummaryView{}
	}
	envelope := impactSummaryEnvelope{}
	if err := json.Unmarshal([]byte(trimmed), &envelope); err != nil {
		return impactSummaryView{}
	}
	return envelope.ImpactSummary
}
