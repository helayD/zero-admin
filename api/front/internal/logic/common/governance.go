package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/cms/cmsclient"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
)

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

func ResolveEffectiveGovernanceScope(ctx context.Context) pkgscope.GovernanceScope {
	current, err := pkgscope.NormalizeGovernanceScope(
		readContextString(ctx, "scopeType"),
		readContextInt64(ctx, "platformId"),
		readContextInt64(ctx, "tenantId"),
		readContextInt64(ctx, "merchantId"),
	)
	if err == nil {
		return current
	}

	defaultScope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypePlatform, pkgscope.DefaultPlatformID, 0, 0)
	return defaultScope
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

func readContextString(ctx context.Context, key string) string {
	value, _ := ctx.Value(key).(string)
	return value
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
