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
