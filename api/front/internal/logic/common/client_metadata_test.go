package common

import (
	"context"
	"net/http/httptest"
	"testing"
)

func TestReadClientRequestMetadata(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/api/app/version/policy", nil)
	req.Header.Set("X-App-Version", " 1.2.0 ")
	req.Header.Set("X-Client-Platform", " android ")
	req.Header.Set("X-Intent-Source", " message_tap ")
	req.Header.Set("X-Intent-Id", " member_message:7 ")
	req.Header.Set("X-Network-State", " wifi ")

	metadata := ReadClientRequestMetadata(req)
	if metadata.AppVersion != "1.2.0" {
		t.Fatalf("expected app version 1.2.0, got %s", metadata.AppVersion)
	}
	if metadata.Platform != "android" {
		t.Fatalf("expected platform android, got %s", metadata.Platform)
	}
	if metadata.IntentSource != "message_tap" {
		t.Fatalf("expected intent source message_tap, got %s", metadata.IntentSource)
	}
	if metadata.IntentID != "member_message:7" {
		t.Fatalf("expected intent id member_message:7, got %s", metadata.IntentID)
	}
	if metadata.NetworkState != "wifi" {
		t.Fatalf("expected network state wifi, got %s", metadata.NetworkState)
	}
}

func TestWithClientRequestMetadata(t *testing.T) {
	t.Parallel()

	ctx := WithClientRequestMetadata(context.Background(), ClientRequestMetadata{
		AppVersion:   "1.1.0",
		Platform:     "ios",
		IntentSource: "bootstrap",
		IntentID:     "upgrade:1",
		NetworkState: "5g",
	})
	metadata := ClientRequestMetadataFromContext(ctx)
	if metadata.AppVersion != "1.1.0" || metadata.Platform != "ios" {
		t.Fatalf("unexpected metadata: %+v", metadata)
	}
}

func TestCompareAppVersion(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		current  string
		required string
		expected int
	}{
		{name: "smaller version", current: "1.0.0", required: "1.2.0", expected: -1},
		{name: "equal with missing patch", current: "1.2", required: "1.2.0", expected: 0},
		{name: "prefixed numeric", current: "1.2.0+4", required: "1.2.0", expected: 0},
		{name: "invalid version", current: "", required: "1.0.0", expected: -1},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := CompareAppVersion(tc.current, tc.required); got != tc.expected {
				t.Fatalf("CompareAppVersion(%q, %q) = %d, want %d", tc.current, tc.required, got, tc.expected)
			}
		})
	}
}
