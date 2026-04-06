// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package version

import (
	"context"
	"slices"
	"strings"
	"time"

	logiccommon "github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"

	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

type QueryAppVersionPolicyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryAppVersionPolicyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryAppVersionPolicyLogic {
	return &QueryAppVersionPolicyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryAppVersionPolicyLogic) QueryAppVersionPolicy(req *types.AppVersionPolicyReq) (resp *types.AppVersionPolicyResp, err error) {
	metadata := logiccommon.ClientRequestMetadataFromContext(l.ctx)
	scene := strings.TrimSpace(req.Scene)
	if scene == "" {
		scene = "app_bootstrap"
	}
	targetType := strings.TrimSpace(req.TargetType)
	affectedCapabilities := normalizeAffectedCapabilities(
		l.svcCtx.Config.UpgradePolicy.AffectedCapabilities,
	)
	policyApplies := shouldApplyUpgradePolicy(
		scene,
		targetType,
		req.Channel,
		affectedCapabilities,
	)

	currentVersion := strings.TrimSpace(metadata.AppVersion)
	if currentVersion == "" {
		currentVersion = strings.TrimSpace(l.svcCtx.Config.UpgradePolicy.CurrentVersion)
	}

	minSupportedVersion := strings.TrimSpace(l.svcCtx.Config.UpgradePolicy.MinSupportedVersion)
	recommendedVersion := strings.TrimSpace(l.svcCtx.Config.UpgradePolicy.RecommendedVersion)
	requiredVersion := ""
	updateMode := "none"
	blocking := false

	if policyApplies {
		switch {
		case minSupportedVersion != "" && logiccommon.CompareAppVersion(currentVersion, minSupportedVersion) < 0:
			requiredVersion = minSupportedVersion
			updateMode = "force"
			blocking = true
		case recommendedVersion != "" && logiccommon.CompareAppVersion(currentVersion, recommendedVersion) < 0:
			requiredVersion = recommendedVersion
			if strings.TrimSpace(l.svcCtx.Config.UpgradePolicy.DeadlineAt) != "" {
				updateMode = "deadline"
			} else {
				updateMode = "recommended"
			}
		}
	}

	traceID := "upgrade-" + time.Now().Format("20060102150405.000000000")
	resp = &types.AppVersionPolicyResp{
		Code:    0,
		Message: "success",
		Data: types.AppVersionPolicyData{
			CurrentVersion:       currentVersion,
			MinSupportedVersion:  minSupportedVersion,
			RecommendedVersion:   recommendedVersion,
			RequiredVersion:      requiredVersion,
			UpdateMode:           updateMode,
			EffectiveAt:          strings.TrimSpace(l.svcCtx.Config.UpgradePolicy.EffectiveAt),
			DeadlineAt:           strings.TrimSpace(l.svcCtx.Config.UpgradePolicy.DeadlineAt),
			AffectedCapabilities: append([]string(nil), affectedCapabilities...),
			Blocking:             blocking,
			ReleaseNotesSummary:  strings.TrimSpace(l.svcCtx.Config.UpgradePolicy.ReleaseNotesSummary),
			UpgradeUrl:           strings.TrimSpace(l.svcCtx.Config.UpgradePolicy.UpgradeUrl),
			StoreTarget:          strings.TrimSpace(l.svcCtx.Config.UpgradePolicy.StoreTarget),
			RecoveryHint:         strings.TrimSpace(l.svcCtx.Config.UpgradePolicy.RecoveryHint),
			TraceId:              traceID,
			Platform:             strings.TrimSpace(metadata.Platform),
			Channel:              strings.TrimSpace(req.Channel),
			InstallerStore:       strings.TrimSpace(req.InstallerStore),
			Scene:                scene,
			TargetType:           targetType,
			TargetId:             req.TargetId,
		},
	}

	logc.Infof(
		l.ctx,
		"查询版本策略, traceId=%s, scene=%s, targetType=%s, targetId=%d, appVersion=%s, platform=%s, blocking=%t, updateMode=%s, requiredVersion=%s, intentSource=%s, intentId=%s, networkState=%s",
		traceID,
		scene,
		resp.Data.TargetType,
		resp.Data.TargetId,
		currentVersion,
		resp.Data.Platform,
		blocking,
		updateMode,
		requiredVersion,
		metadata.IntentSource,
		metadata.IntentID,
		metadata.NetworkState,
	)

	return resp, nil
}

func normalizeAffectedCapabilities(values []string) []string {
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(strings.ToLower(value))
		if trimmed == "" || slices.Contains(normalized, trimmed) {
			continue
		}
		normalized = append(normalized, trimmed)
	}
	return normalized
}

func shouldApplyUpgradePolicy(
	scene, targetType, channel string,
	affectedCapabilities []string,
) bool {
	normalizedScene := strings.TrimSpace(strings.ToLower(scene))
	normalizedTargetType := strings.TrimSpace(strings.ToLower(targetType))
	normalizedChannel := strings.TrimSpace(strings.ToLower(channel))
	if normalizedScene == "settings_check" {
		return true
	}
	if len(affectedCapabilities) == 0 {
		return true
	}
	for _, capability := range affectedCapabilities {
		switch capability {
		case normalizedScene, "scene:" + normalizedScene:
			return true
		case normalizedTargetType, "target_type:" + normalizedTargetType, "targettype:" + normalizedTargetType:
			return true
		case normalizedChannel, "channel:" + normalizedChannel:
			if normalizedChannel != "" {
				return true
			}
		}
	}
	return false
}
