package message

import (
	"context"
	"net/http"

	logiccommon "github.com/feihua/zero-admin/api/front/internal/logic/common"
	logicmessage "github.com/feihua/zero-admin/api/front/internal/logic/member/message"
)

func enrichRecallRequestContext(ctx context.Context, r *http.Request) context.Context {
	metadata := logiccommon.ReadClientRequestMetadata(r)
	return logicmessage.WithRecallRequestMetadata(
		ctx,
		logicmessage.RecallRequestMetadata{
			AppVersion:   metadata.AppVersion,
			Platform:     metadata.Platform,
			IntentSource: metadata.IntentSource,
			IntentID:     metadata.IntentID,
			NetworkState: metadata.NetworkState,
		},
	)
}
