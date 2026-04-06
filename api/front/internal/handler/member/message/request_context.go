package message

import (
	"context"
	"net/http"

	logicmessage "github.com/feihua/zero-admin/api/front/internal/logic/member/message"
)

func enrichRecallRequestContext(ctx context.Context, r *http.Request) context.Context {
	return logicmessage.WithRecallRequestMetadata(
		ctx,
		logicmessage.RecallRequestMetadata{
			AppVersion:   r.Header.Get("X-App-Version"),
			Platform:     r.Header.Get("X-Client-Platform"),
			IntentSource: r.Header.Get("X-Intent-Source"),
			IntentID:     r.Header.Get("X-Intent-Id"),
			NetworkState: r.Header.Get("X-Network-State"),
		},
	)
}
