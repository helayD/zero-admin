package physical_fulfillment

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/feihua/zero-admin/pkg/digitalcardmint"
)

func TestMapPhysicalDetailDoesNotExposeChainFields(t *testing.T) {
	data := mapPhysicalDetail(&digitalcardmint.MemberPhysicalFulfillmentDetail{
		FulfillmentID:         1,
		FulfillmentNo:         "PF2026042600000001",
		AssetInstanceID:       970001,
		AssetNo:               "CARD-001",
		TemplateName:          "SSR 卡",
		ActivityName:          "春季抽卡",
		ObtainedAt:            "2026-04-26 10:00:00",
		MintStatusText:        "已到账",
		FulfillmentStatus:     digitalcardmint.PhysicalFulfillmentStatusShipped,
		FulfillmentStatusText: "已发货",
		ProductionStatusText:  "制作完成",
		ShippingStatusText:    "已发货",
		ReceiverNameMasked:    "张*",
		ReceiverPhoneMasked:   "138****8001",
		AddressSummary:        "广东省深圳市南山区科技园",
		CarrierName:           "顺丰速运",
		TrackingNo:            "SF123456789",
		ComplianceTipSummary:  "实体卡履约仅展示制作、配送和签收进度。",
		Timeline: []digitalcardmint.PhysicalFulfillmentTimelineItem{
			{Action: "physical_shipped", ActionText: "实体卡发货", StatusText: "已发货", Reason: "发货", CreateTime: "2026-04-26 11:00:00"},
		},
	}, nil, 970001)

	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal physical detail failed: %v", err)
	}
	body := string(raw)
	for _, forbidden := range []string{
		"chainType",
		"chainStatus",
		"chainTxId",
		"tokenId",
		"lastReceiptJson",
		"lastReceiptSummary",
		"蚂蚁链",
		"AntChain",
		"FISCO",
		"区块链",
		"链上",
		"上链",
		"链路",
	} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("front physical fulfillment response leaked %q: %s", forbidden, body)
		}
	}
}

func TestMapPhysicalDetailBlockedKeepsAssetSummary(t *testing.T) {
	data := mapPhysicalDetail(nil, &digitalcardmint.PhysicalFulfillmentResult{
		AssetInstanceID:       970009,
		FulfillmentStatus:     digitalcardmint.PhysicalFulfillmentStatusPendingDigitalConfirmation,
		FulfillmentStatusText: "待权益确认",
		ShippingFeeStatus:     digitalcardmint.PhysicalShippingFeeStatusPending,
		ShippingFeeStatusText: "待发放后确认",
		BlockedReason:         digitalcardmint.PhysicalBlockDigitalPending,
		BlockedReasonText:     "待完成权益确认后制作",
	}, 970009, &digitalcardmint.MemberDigitalCardAssetDetail{
		Item: digitalcardmint.MemberDigitalCardAssetItem{
			AssetInstanceID:       970009,
			AssetNo:               "CARD20260426085756FB38DA19",
			TemplateName:          "SR 星云狐",
			ActivityName:          "2026 春季限定数字卡片抽卡",
			ObtainedAt:            "2026-04-26 08:57:56",
			MintStatusText:        "待发放",
			ComplianceRuleSummary: "默认禁止收益承诺",
		},
	})

	if data.AssetNo == "" || data.TemplateName == "" || data.MintStatusText == "" {
		t.Fatalf("blocked physical detail lost asset summary: %+v", data)
	}
	if data.BlockedReasonText != "待完成权益确认后制作" {
		t.Fatalf("unexpected blocked reason text: %s", data.BlockedReasonText)
	}
	if data.ShippingFeeStatusText != "待发放后确认" {
		t.Fatalf("unexpected shipping fee status text: %s", data.ShippingFeeStatusText)
	}
}
