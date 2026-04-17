package draw_activity

import (
	"testing"

	"github.com/feihua/zero-admin/rpc/sms/smsclient"
)

func TestMapDrawRecordIncludesAssetFields(t *testing.T) {
	record := mapDrawRecord(&smsclient.DrawMemberRecordData{
		Id:               11,
		ActivityId:       22,
		RequestId:        "req-front-1",
		ResultType:       "won",
		ResultStatus:     "won_pending_asset",
		ResultStatusText: "已中奖待到账",
		TemplateName:     "SSR 卡",
		Rarity:           "SSR",
		AssetInstanceId:  1001,
		AssetNo:          "CARD202604170001",
		AssetStatus:      "asset_created",
		AssetStatusText:  "资产已创建，链上处理中",
		AssetCreatedAt:   "2026-04-17 10:00:00",
	})

	if record.AssetInstanceId != 1001 {
		t.Fatalf("expected assetInstanceId 1001, got %d", record.AssetInstanceId)
	}
	if record.AssetNo != "CARD202604170001" {
		t.Fatalf("expected assetNo to be mapped, got %s", record.AssetNo)
	}
	if record.AssetStatusText != "资产已创建，链上处理中" {
		t.Fatalf("expected assetStatusText to be mapped, got %s", record.AssetStatusText)
	}
	if record.AssetCreatedAt != "2026-04-17 10:00:00" {
		t.Fatalf("expected assetCreatedAt to be mapped, got %s", record.AssetCreatedAt)
	}
}

func TestMapDrawRecordKeepsEmptyAssetNumberWhenServerDoesNotProvideIt(t *testing.T) {
	record := mapDrawRecord(&smsclient.DrawMemberRecordData{
		Id:              12,
		ActivityId:      22,
		RequestId:       "req-front-2",
		ResultType:      "won",
		ResultStatus:    "won_pending_asset",
		AssetInstanceId: 1002,
		AssetNo:         "",
		AssetStatus:     "asset_created",
		AssetStatusText: "资产已创建，链上处理中",
	})

	if record.AssetInstanceId != 1002 {
		t.Fatalf("expected assetInstanceId 1002, got %d", record.AssetInstanceId)
	}
	if record.AssetNo != "" {
		t.Fatalf("expected empty assetNo to be preserved, got %s", record.AssetNo)
	}
	if record.AssetStatusText != "资产已创建，链上处理中" {
		t.Fatalf("expected assetStatusText to be mapped, got %s", record.AssetStatusText)
	}
}
