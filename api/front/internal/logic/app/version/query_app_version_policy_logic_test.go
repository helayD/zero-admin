package version

import (
	"context"
	"testing"

	"github.com/feihua/zero-admin/api/front/internal/config"
	logiccommon "github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
)

func TestQueryAppVersionPolicyReturnsForceModeWhenVersionBelowMinimum(t *testing.T) {
	t.Parallel()

	var cfg config.Config
	cfg.UpgradePolicy.CurrentVersion = "1.0.0"
	cfg.UpgradePolicy.MinSupportedVersion = "1.2.0"
	cfg.UpgradePolicy.RecommendedVersion = "1.3.0"
	cfg.UpgradePolicy.UpgradeUrl = "https://example.com/app.apk"
	cfg.UpgradePolicy.StoreTarget = "browser_download"
	cfg.UpgradePolicy.RecoveryHint = "升级后可恢复原任务"
	cfg.UpgradePolicy.AffectedCapabilities = []string{"order_pay"}

	ctx := logiccommon.WithClientRequestMetadata(context.Background(), logiccommon.ClientRequestMetadata{
		AppVersion: "1.0.0",
		Platform:   "android",
	})
	logic := NewQueryAppVersionPolicyLogic(ctx, &svc.ServiceContext{Config: cfg})

	resp, err := logic.QueryAppVersionPolicy(&types.AppVersionPolicyReq{
		Scene:      "order_pay",
		TargetType: "order_detail",
		TargetId:   9001,
		Channel:    "direct",
	})
	if err != nil {
		t.Fatalf("QueryAppVersionPolicy returned error: %v", err)
	}
	if !resp.Data.Blocking {
		t.Fatalf("expected blocking policy, got %+v", resp.Data)
	}
	if resp.Data.UpdateMode != "force" {
		t.Fatalf("expected force mode, got %s", resp.Data.UpdateMode)
	}
	if resp.Data.RequiredVersion != "1.2.0" {
		t.Fatalf("expected required version 1.2.0, got %s", resp.Data.RequiredVersion)
	}
}

func TestQueryAppVersionPolicyReturnsDeadlineModeForRecommendedUpgrade(t *testing.T) {
	t.Parallel()

	var cfg config.Config
	cfg.UpgradePolicy.CurrentVersion = "1.3.0"
	cfg.UpgradePolicy.MinSupportedVersion = "1.0.0"
	cfg.UpgradePolicy.RecommendedVersion = "1.3.0"
	cfg.UpgradePolicy.DeadlineAt = "2026-05-31T23:59:59+08:00"
	cfg.UpgradePolicy.RecoveryHint = "升级后可恢复原任务"

	ctx := logiccommon.WithClientRequestMetadata(context.Background(), logiccommon.ClientRequestMetadata{
		AppVersion: "1.2.0",
		Platform:   "ios",
	})
	logic := NewQueryAppVersionPolicyLogic(ctx, &svc.ServiceContext{Config: cfg})

	resp, err := logic.QueryAppVersionPolicy(&types.AppVersionPolicyReq{
		Scene:      "settings_check",
		TargetType: "settings",
	})
	if err != nil {
		t.Fatalf("QueryAppVersionPolicy returned error: %v", err)
	}
	if resp.Data.Blocking {
		t.Fatalf("expected non-blocking policy, got %+v", resp.Data)
	}
	if resp.Data.UpdateMode != "deadline" {
		t.Fatalf("expected deadline mode, got %s", resp.Data.UpdateMode)
	}
	if resp.Data.CurrentVersion != "1.2.0" {
		t.Fatalf("expected current version 1.2.0, got %s", resp.Data.CurrentVersion)
	}
}

func TestQueryAppVersionPolicySkipsUnmatchedCapabilityScene(t *testing.T) {
	t.Parallel()

	var cfg config.Config
	cfg.UpgradePolicy.CurrentVersion = "1.0.0"
	cfg.UpgradePolicy.MinSupportedVersion = "1.2.0"
	cfg.UpgradePolicy.RecommendedVersion = "1.3.0"
	cfg.UpgradePolicy.UpgradeUrl = "https://example.com/app.apk"
	cfg.UpgradePolicy.StoreTarget = "browser_download"
	cfg.UpgradePolicy.AffectedCapabilities = []string{"order_pay"}

	ctx := logiccommon.WithClientRequestMetadata(context.Background(), logiccommon.ClientRequestMetadata{
		AppVersion: "1.0.0",
		Platform:   "android",
	})
	logic := NewQueryAppVersionPolicyLogic(ctx, &svc.ServiceContext{Config: cfg})

	resp, err := logic.QueryAppVersionPolicy(&types.AppVersionPolicyReq{
		Scene:      "app_bootstrap",
		TargetType: "home",
	})
	if err != nil {
		t.Fatalf("QueryAppVersionPolicy returned error: %v", err)
	}
	if resp.Data.Blocking {
		t.Fatalf("expected unmatched capability to skip blocking policy, got %+v", resp.Data)
	}
	if resp.Data.UpdateMode != "none" {
		t.Fatalf("expected unmatched capability to return none mode, got %s", resp.Data.UpdateMode)
	}
	if resp.Data.RequiredVersion != "" {
		t.Fatalf("expected unmatched capability to return empty required version, got %s", resp.Data.RequiredVersion)
	}
}

func TestQueryAppVersionPolicyAlwaysAllowsSettingsManualCheck(t *testing.T) {
	t.Parallel()

	var cfg config.Config
	cfg.UpgradePolicy.CurrentVersion = "1.0.0"
	cfg.UpgradePolicy.MinSupportedVersion = "1.0.0"
	cfg.UpgradePolicy.RecommendedVersion = "1.3.0"
	cfg.UpgradePolicy.DeadlineAt = "2026-05-31T23:59:59+08:00"
	cfg.UpgradePolicy.AffectedCapabilities = []string{"order_pay"}

	ctx := logiccommon.WithClientRequestMetadata(context.Background(), logiccommon.ClientRequestMetadata{
		AppVersion: "1.2.0",
		Platform:   "android",
	})
	logic := NewQueryAppVersionPolicyLogic(ctx, &svc.ServiceContext{Config: cfg})

	resp, err := logic.QueryAppVersionPolicy(&types.AppVersionPolicyReq{
		Scene:      "settings_check",
		TargetType: "settings",
	})
	if err != nil {
		t.Fatalf("QueryAppVersionPolicy returned error: %v", err)
	}
	if resp.Data.UpdateMode != "deadline" {
		t.Fatalf("expected settings check to surface manual update policy, got %s", resp.Data.UpdateMode)
	}
	if resp.Data.Blocking {
		t.Fatalf("expected settings check recommendation to stay non-blocking, got %+v", resp.Data)
	}
}

func TestQueryAppVersionPolicyMatchesChannelCapability(t *testing.T) {
	t.Parallel()

	var cfg config.Config
	cfg.UpgradePolicy.CurrentVersion = "1.0.0"
	cfg.UpgradePolicy.MinSupportedVersion = "1.2.0"
	cfg.UpgradePolicy.RecommendedVersion = "1.3.0"
	cfg.UpgradePolicy.UpgradeUrl = "https://example.com/app.apk"
	cfg.UpgradePolicy.StoreTarget = "browser_download"
	cfg.UpgradePolicy.AffectedCapabilities = []string{"channel:gray"}

	ctx := logiccommon.WithClientRequestMetadata(context.Background(), logiccommon.ClientRequestMetadata{
		AppVersion: "1.0.0",
		Platform:   "android",
	})
	logic := NewQueryAppVersionPolicyLogic(ctx, &svc.ServiceContext{Config: cfg})

	resp, err := logic.QueryAppVersionPolicy(&types.AppVersionPolicyReq{
		Scene:   "app_bootstrap",
		Channel: "gray",
	})
	if err != nil {
		t.Fatalf("QueryAppVersionPolicy returned error: %v", err)
	}
	if !resp.Data.Blocking {
		t.Fatalf("expected channel capability to match blocking policy, got %+v", resp.Data)
	}
	if resp.Data.UpdateMode != "force" {
		t.Fatalf("expected force mode when channel capability matches, got %s", resp.Data.UpdateMode)
	}
}
