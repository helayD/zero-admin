package message

import (
	"context"
	"strings"

	logiccommon "github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/rpc/ums/umsclient"
	"github.com/zeromicro/go-zero/core/logc"
)

type RecallRequestMetadata struct {
	AppVersion   string
	Platform     string
	IntentSource string
	IntentID     string
	NetworkState string
}

type recallRequestContextKey string

const (
	recallAppVersionContextKey   recallRequestContextKey = "message_recall_app_version"
	recallPlatformContextKey     recallRequestContextKey = "message_recall_platform"
	recallIntentSourceContextKey recallRequestContextKey = "message_recall_intent_source"
	recallIntentIDContextKey     recallRequestContextKey = "message_recall_intent_id"
	recallNetworkStateContextKey recallRequestContextKey = "message_recall_network_state"
)

const (
	recallFailureMinVersionUnmet = "min_version_unmet"
)

func WithRecallRequestMetadata(ctx context.Context, metadata RecallRequestMetadata) context.Context {
	ctx = logiccommon.WithClientRequestMetadata(ctx, logiccommon.ClientRequestMetadata{
		AppVersion:   metadata.AppVersion,
		Platform:     metadata.Platform,
		IntentSource: metadata.IntentSource,
		IntentID:     metadata.IntentID,
		NetworkState: metadata.NetworkState,
	})
	ctx = context.WithValue(ctx, recallAppVersionContextKey, strings.TrimSpace(metadata.AppVersion))
	ctx = context.WithValue(ctx, recallPlatformContextKey, strings.TrimSpace(metadata.Platform))
	ctx = context.WithValue(ctx, recallIntentSourceContextKey, strings.TrimSpace(metadata.IntentSource))
	ctx = context.WithValue(ctx, recallIntentIDContextKey, strings.TrimSpace(metadata.IntentID))
	ctx = context.WithValue(ctx, recallNetworkStateContextKey, strings.TrimSpace(metadata.NetworkState))
	return ctx
}

func messageIntentResponseFromRPC(ctx context.Context, item *umsclient.MemberMessageData) *types.MemberMessageIntentResp {
	if item == nil || item.Intent == nil {
		return nil
	}

	intent := &types.MemberMessageIntentResp{
		IntentType:       item.Intent.IntentType,
		TargetType:       item.Intent.TargetType,
		TargetID:         recallIDPointer(item.Intent.TargetId),
		TargetTab:        recallTabPointer(item.Intent.TargetType, item.Intent.TargetTab),
		FallbackType:     item.Intent.FallbackType,
		FallbackTargetID: recallIDPointer(item.Intent.FallbackTargetId),
		FallbackTab:      recallTabPointer(item.Intent.FallbackType, item.Intent.FallbackTab),
		RequiresAuth:     item.Intent.RequiresAuth,
		MinAppVersion:    item.Intent.MinAppVersion,
		IntentID:         item.Intent.IntentId,
		IssuedAt:         item.Intent.IssuedAt,
		FailureReason:    item.Intent.FailureReason,
		RecoveryHint:     item.Intent.RecoveryHint,
		Blocked:          item.Intent.Blocked,
		Source:           item.Intent.Source,
	}

	if appVersion := recallStringFromContext(ctx, recallAppVersionContextKey); appVersion != "" &&
		intent.MinAppVersion != "" &&
		logiccommon.CompareAppVersion(appVersion, intent.MinAppVersion) < 0 {
		intent.Blocked = true
		intent.FailureReason = recallFailureMinVersionUnmet
		if strings.TrimSpace(intent.RecoveryHint) == "" {
			intent.RecoveryHint = "当前客户端版本暂不支持直达该入口，已为你返回可用页面"
		}
		logc.Infof(
			ctx,
			"消息唤回因客户端版本不足被阻断, intentId=%s, targetType=%s, appVersion=%s, minVersion=%s, source=%s, platform=%s, networkState=%s",
			intent.IntentID,
			intent.TargetType,
			appVersion,
			intent.MinAppVersion,
			defaultRecallValue(intent.Source, recallStringFromContext(ctx, recallIntentSourceContextKey)),
			recallStringFromContext(ctx, recallPlatformContextKey),
			recallStringFromContext(ctx, recallNetworkStateContextKey),
		)
	}

	return intent
}

func recallStringFromContext(ctx context.Context, key recallRequestContextKey) string {
	value, _ := ctx.Value(key).(string)
	return strings.TrimSpace(value)
}

func defaultRecallValue(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func recallIDPointer(value int64) *int64 {
	if value <= 0 {
		return nil
	}
	return &value
}

func recallTabPointer(targetType string, value int32) *int64 {
	normalizedTargetType := strings.TrimSpace(strings.ToLower(targetType))
	if !recallTargetSupportsTab(normalizedTargetType) {
		return nil
	}
	if value == 0 {
		switch normalizedTargetType {
		case "order_list", "coupon_list":
		default:
			return nil
		}
	}
	result := int64(value)
	return &result
}

func recallTargetSupportsTab(targetType string) bool {
	switch strings.TrimSpace(strings.ToLower(targetType)) {
	case "home", "cart", "order_list", "coupon_list":
		return true
	default:
		return false
	}
}
