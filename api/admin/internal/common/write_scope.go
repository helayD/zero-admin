package common

import (
	"context"
	"errors"
	"fmt"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
)

func ResolveWriteGovernanceScope(ctx context.Context, requested RequestedGovernanceScope) (pkgscope.GovernanceScope, error) {
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
	if err := ensureWriteScopeAllowed(current, resolved); err != nil {
		return pkgscope.GovernanceScope{}, err
	}

	return resolved, nil
}

func ensureWriteScopeAllowed(current, requested pkgscope.GovernanceScope) error {
	switch current.ScopeType {
	case pkgscope.SubjectTypePlatform:
		return nil
	case pkgscope.SubjectTypeTenant, pkgscope.SubjectTypeMerchant:
		if current.SameScope(requested) {
			return nil
		}
		return errors.New("当前主体不允许切换写入范围")
	default:
		return fmt.Errorf("不支持的当前主体范围：%s", current.ScopeType)
	}
}
