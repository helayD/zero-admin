package common

import (
	"context"
	"net/http"
	"strings"
)

type ClientRequestMetadata struct {
	AppVersion   string
	Platform     string
	IntentSource string
	IntentID     string
	NetworkState string
}

type clientRequestMetadataKey string

const clientRequestMetadataContextKey clientRequestMetadataKey = "front_client_request_metadata"

func ReadClientRequestMetadata(r *http.Request) ClientRequestMetadata {
	if r == nil {
		return ClientRequestMetadata{}
	}
	return ClientRequestMetadata{
		AppVersion:   strings.TrimSpace(r.Header.Get("X-App-Version")),
		Platform:     strings.TrimSpace(r.Header.Get("X-Client-Platform")),
		IntentSource: strings.TrimSpace(r.Header.Get("X-Intent-Source")),
		IntentID:     strings.TrimSpace(r.Header.Get("X-Intent-Id")),
		NetworkState: strings.TrimSpace(r.Header.Get("X-Network-State")),
	}
}

func WithClientRequestMetadata(ctx context.Context, metadata ClientRequestMetadata) context.Context {
	return context.WithValue(ctx, clientRequestMetadataContextKey, ClientRequestMetadata{
		AppVersion:   strings.TrimSpace(metadata.AppVersion),
		Platform:     strings.TrimSpace(metadata.Platform),
		IntentSource: strings.TrimSpace(metadata.IntentSource),
		IntentID:     strings.TrimSpace(metadata.IntentID),
		NetworkState: strings.TrimSpace(metadata.NetworkState),
	})
}

func ClientRequestMetadataFromContext(ctx context.Context) ClientRequestMetadata {
	if ctx == nil {
		return ClientRequestMetadata{}
	}
	metadata, _ := ctx.Value(clientRequestMetadataContextKey).(ClientRequestMetadata)
	return ClientRequestMetadata{
		AppVersion:   strings.TrimSpace(metadata.AppVersion),
		Platform:     strings.TrimSpace(metadata.Platform),
		IntentSource: strings.TrimSpace(metadata.IntentSource),
		IntentID:     strings.TrimSpace(metadata.IntentID),
		NetworkState: strings.TrimSpace(metadata.NetworkState),
	}
}

func CompareAppVersion(current, required string) int {
	currentParts := SplitAppVersion(current)
	requiredParts := SplitAppVersion(required)
	maxLen := len(currentParts)
	if len(requiredParts) > maxLen {
		maxLen = len(requiredParts)
	}

	for i := 0; i < maxLen; i++ {
		currentValue := versionPartAt(currentParts, i)
		requiredValue := versionPartAt(requiredParts, i)
		if currentValue > requiredValue {
			return 1
		}
		if currentValue < requiredValue {
			return -1
		}
	}

	return 0
}

func SplitAppVersion(value string) []int {
	rawParts := strings.Split(strings.TrimSpace(value), ".")
	result := make([]int, 0, len(rawParts))
	for _, rawPart := range rawParts {
		number := 0
		for _, char := range strings.TrimSpace(rawPart) {
			if char < '0' || char > '9' {
				break
			}
			number = number*10 + int(char-'0')
		}
		result = append(result, number)
	}
	return result
}

func versionPartAt(parts []int, idx int) int {
	if idx < 0 || idx >= len(parts) {
		return 0
	}
	return parts[idx]
}
