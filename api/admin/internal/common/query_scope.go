package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/cms/cmsclient"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
)

type RequestedGovernanceScope struct {
	ScopeType  string
	PlatformID int64
	TenantID   int64
	MerchantID int64
}

func CurrentGovernanceScope(ctx context.Context) (pkgscope.GovernanceScope, error) {
	scopeType, _ := ctx.Value("scopeType").(string)
	if strings.TrimSpace(scopeType) == "" {
		return pkgscope.GovernanceScope{}, errors.New("当前登录上下文缺少治理范围")
	}

	current, err := pkgscope.NormalizeGovernanceScope(
		scopeType,
		readContextInt64(ctx, "platformId"),
		readContextInt64(ctx, "tenantId"),
		readContextInt64(ctx, "merchantId"),
	)
	if err != nil {
		return pkgscope.GovernanceScope{}, fmt.Errorf("当前登录上下文治理范围无效: %w", err)
	}

	return current, nil
}

func ResolveQueryGovernanceScope(ctx context.Context, requested RequestedGovernanceScope) (pkgscope.GovernanceScope, error) {
	current, err := CurrentGovernanceScope(ctx)
	if err != nil {
		return pkgscope.GovernanceScope{}, err
	}
	if !requested.provided() {
		return current, nil
	}

	resolved, err := requested.normalizeAgainst(current)
	if err != nil {
		return pkgscope.GovernanceScope{}, err
	}
	if err := ensureQueryScopeAllowed(current, resolved); err != nil {
		return pkgscope.GovernanceScope{}, err
	}

	return resolved, nil
}

func PMSGovernanceScope(current pkgscope.GovernanceScope) *pmsclient.GovernanceScope {
	return &pmsclient.GovernanceScope{
		ScopeType:  current.ScopeType,
		PlatformId: current.PlatformID,
		TenantId:   current.TenantID,
		MerchantId: current.MerchantID,
		ScopeLabel: current.Label(),
	}
}

func OMSGovernanceScope(current pkgscope.GovernanceScope) *omsclient.GovernanceScope {
	return &omsclient.GovernanceScope{
		ScopeType:  current.ScopeType,
		PlatformId: current.PlatformID,
		TenantId:   current.TenantID,
		MerchantId: current.MerchantID,
		ScopeLabel: current.Label(),
	}
}

func SMSGovernanceScope(current pkgscope.GovernanceScope) *smsclient.GovernanceScope {
	return &smsclient.GovernanceScope{
		ScopeType:  current.ScopeType,
		PlatformId: current.PlatformID,
		TenantId:   current.TenantID,
		MerchantId: current.MerchantID,
		ScopeLabel: current.Label(),
	}
}

func CMSGovernanceScope(current pkgscope.GovernanceScope) *cmsclient.GovernanceScope {
	return &cmsclient.GovernanceScope{
		ScopeType:  current.ScopeType,
		PlatformId: current.PlatformID,
		TenantId:   current.TenantID,
		MerchantId: current.MerchantID,
		ScopeLabel: current.Label(),
	}
}

func ensureQueryScopeAllowed(current, requested pkgscope.GovernanceScope) error {
	switch current.ScopeType {
	case pkgscope.SubjectTypePlatform:
		return nil
	case pkgscope.SubjectTypeTenant, pkgscope.SubjectTypeMerchant:
		if current.SameScope(requested) {
			return nil
		}
		return errors.New("当前主体不允许切换查询范围")
	default:
		return fmt.Errorf("不支持的当前主体范围：%s", current.ScopeType)
	}
}

func (r RequestedGovernanceScope) provided() bool {
	return strings.TrimSpace(r.ScopeType) != "" || r.PlatformID > 0 || r.TenantID > 0 || r.MerchantID > 0
}

func (r RequestedGovernanceScope) normalizeAgainst(current pkgscope.GovernanceScope) (pkgscope.GovernanceScope, error) {
	scopeType := strings.TrimSpace(r.ScopeType)
	platformID := r.PlatformID
	tenantID := r.TenantID
	merchantID := r.MerchantID

	if platformID == 0 {
		platformID = current.PlatformID
	}
	if scopeType == "" || scopeType == current.ScopeType {
		if tenantID == 0 {
			tenantID = current.TenantID
		}
		if merchantID == 0 {
			merchantID = current.MerchantID
		}
	}

	return pkgscope.NormalizeGovernanceScope(scopeType, platformID, tenantID, merchantID)
}

func readContextInt64(ctx context.Context, key string) int64 {
	value := ctx.Value(key)
	switch typed := value.(type) {
	case json.Number:
		number, _ := typed.Int64()
		return number
	case float64:
		return int64(typed)
	case float32:
		return int64(typed)
	case int64:
		return typed
	case int32:
		return int64(typed)
	case int:
		return int64(typed)
	default:
		return 0
	}
}
