package digitalcardmint

import (
	"context"
	"fmt"
	"strings"
	"testing"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"gorm.io/gorm"
)

func preparePhysicalFulfillmentTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db := newDigitalCardMintTestDB(t)
	stmts := []string{
		`CREATE TABLE ums_member_address (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			member_id INTEGER NOT NULL,
			receiver_name TEXT NOT NULL DEFAULT '',
			receiver_phone TEXT NOT NULL DEFAULT '',
			province TEXT NOT NULL DEFAULT '',
			city TEXT NOT NULL DEFAULT '',
			district TEXT NOT NULL DEFAULT '',
			detail_address TEXT NOT NULL DEFAULT '',
			postal_code TEXT NOT NULL DEFAULT '',
			is_default INTEGER NOT NULL DEFAULT 0,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE sms_card_physical_fulfillment (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			platform_id INTEGER NOT NULL DEFAULT 0,
			tenant_id INTEGER NOT NULL DEFAULT 0,
			merchant_id INTEGER NOT NULL DEFAULT 0,
			asset_instance_id INTEGER NOT NULL,
			participation_record_id INTEGER NOT NULL DEFAULT 0,
			activity_id INTEGER NOT NULL DEFAULT 0,
			template_id INTEGER NOT NULL DEFAULT 0,
			member_id INTEGER NOT NULL DEFAULT 0,
			asset_no TEXT NOT NULL DEFAULT '',
			fulfillment_no TEXT NOT NULL DEFAULT '',
			fulfillment_status TEXT NOT NULL DEFAULT '',
			production_status TEXT NOT NULL DEFAULT '',
			quality_status TEXT NOT NULL DEFAULT '',
			shipping_status TEXT NOT NULL DEFAULT '',
			address_id INTEGER NOT NULL DEFAULT 0,
			receiver_name TEXT NOT NULL DEFAULT '',
			receiver_phone TEXT NOT NULL DEFAULT '',
			receiver_province TEXT NOT NULL DEFAULT '',
			receiver_city TEXT NOT NULL DEFAULT '',
			receiver_district TEXT NOT NULL DEFAULT '',
			receiver_address TEXT NOT NULL DEFAULT '',
			receiver_postal_code TEXT NOT NULL DEFAULT '',
			production_batch_no TEXT NOT NULL DEFAULT '',
			carrier_code TEXT NOT NULL DEFAULT '',
			carrier_name TEXT NOT NULL DEFAULT '',
			tracking_no TEXT NOT NULL DEFAULT '',
			shipped_at DATETIME NULL,
			signed_at DATETIME NULL,
			auto_sign_at DATETIME NULL,
			failure_code TEXT NOT NULL DEFAULT '',
			failure_reason TEXT NOT NULL DEFAULT '',
			operator_id INTEGER NOT NULL DEFAULT 0,
			trace_id TEXT NOT NULL DEFAULT '',
			request_id TEXT NOT NULL DEFAULT '',
			create_by INTEGER NOT NULL DEFAULT 0,
			update_by INTEGER NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NULL,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE UNIQUE INDEX uk_physical_asset ON sms_card_physical_fulfillment(asset_instance_id, is_deleted)`,
		`CREATE UNIQUE INDEX uk_physical_no ON sms_card_physical_fulfillment(fulfillment_no, is_deleted)`,
		`CREATE UNIQUE INDEX uk_physical_tracking ON sms_card_physical_fulfillment(tracking_no, carrier_code, is_deleted) WHERE tracking_no <> '' AND carrier_code <> ''`,
		`CREATE TABLE sms_card_physical_fulfillment_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			fulfillment_id INTEGER NOT NULL,
			asset_instance_id INTEGER NOT NULL,
			action TEXT NOT NULL DEFAULT '',
			from_status TEXT NOT NULL DEFAULT '',
			to_status TEXT NOT NULL DEFAULT '',
			operator_type TEXT NOT NULL DEFAULT '',
			operator_id INTEGER NOT NULL DEFAULT 0,
			reason TEXT NOT NULL DEFAULT '',
			payload_json TEXT NOT NULL DEFAULT '',
			trace_id TEXT NOT NULL DEFAULT '',
			request_id TEXT NOT NULL DEFAULT '',
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("exec physical schema failed: %v", err)
		}
	}

	if err := db.Exec(`
		INSERT INTO ums_member_address
			(id, member_id, receiver_name, receiver_phone, province, city, district, detail_address, postal_code, is_default, is_deleted)
		VALUES
			(501, 3001, '张三', '13800138001', '广东省', '深圳市', '南山区', '科技园科兴科学园B座', '518057', 1, 0)
	`).Error; err != nil {
		t.Fatalf("seed address failed: %v", err)
	}
	if err := db.Table(CardInstanceRow{}.TableName()).
		Where("id = 1").
		Updates(map[string]interface{}{
			"mint_status":  MintStatusSuccess,
			"chain_status": ChainStatusSuccess,
			"token_id":     "free-token-001",
		}).Error; err != nil {
		t.Fatalf("prepare minted instance failed: %v", err)
	}
	return db
}

func physicalTestScope() pkgscope.GovernanceScope {
	return pkgscope.GovernanceScope{
		ScopeType:  pkgscope.SubjectTypeMerchant,
		PlatformID: 1,
		TenantID:   10,
		MerchantID: 88,
	}
}

func TestEnsurePhysicalFulfillmentByAssetCreatesPendingAddressOnce(t *testing.T) {
	db := preparePhysicalFulfillmentTestDB(t)
	service := NewService(db, nil, nil)

	first, err := service.EnsurePhysicalFulfillmentByAsset(context.Background(), physicalTestScope(), PhysicalFulfillmentInput{
		AssetInstanceID: 1,
		OperatorType:    OperatorSystem,
		TraceID:         "trace-physical-1",
		RequestID:       "req-physical-1",
	})
	if err != nil {
		t.Fatalf("EnsurePhysicalFulfillmentByAsset returned error: %v", err)
	}
	if first.BlockedReason != "" {
		t.Fatalf("expected no block reason, got %+v", first)
	}
	if first.FulfillmentStatus != PhysicalFulfillmentStatusPendingAddress {
		t.Fatalf("expected pending address status, got %+v", first)
	}

	second, err := service.EnsurePhysicalFulfillmentByAsset(context.Background(), physicalTestScope(), PhysicalFulfillmentInput{
		AssetInstanceID: 1,
		OperatorType:    OperatorJob,
	})
	if err != nil {
		t.Fatalf("second EnsurePhysicalFulfillmentByAsset returned error: %v", err)
	}
	if second.FulfillmentID != first.FulfillmentID || second.FulfillmentNo != first.FulfillmentNo {
		t.Fatalf("expected idempotent fulfillment, first=%+v second=%+v", first, second)
	}
	if !strings.HasPrefix(first.FulfillmentNo, "PF") {
		t.Fatalf("expected generated fulfillment no, got %s", first.FulfillmentNo)
	}

	var count int64
	if err := db.Table(PhysicalFulfillmentRow{}.TableName()).Count(&count).Error; err != nil {
		t.Fatalf("count fulfillment failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one fulfillment row, got %d", count)
	}
}

func TestEnsurePhysicalFulfillmentBlocksDigitalPendingAndRealName(t *testing.T) {
	db := preparePhysicalFulfillmentTestDB(t)
	service := NewService(db, nil, nil)

	if err := db.Table(CardInstanceRow{}.TableName()).
		Where("id = 1").
		Updates(map[string]interface{}{
			"mint_status":  MintStatusPending,
			"chain_status": ChainStatusUnknown,
			"token_id":     "",
		}).Error; err != nil {
		t.Fatalf("prepare pending asset failed: %v", err)
	}
	result, err := service.EnsurePhysicalFulfillmentByAsset(context.Background(), physicalTestScope(), PhysicalFulfillmentInput{AssetInstanceID: 1})
	if err != nil {
		t.Fatalf("EnsurePhysicalFulfillmentByAsset pending returned error: %v", err)
	}
	if result.BlockedReason != PhysicalBlockDigitalPending {
		t.Fatalf("expected digital pending block, got %+v", result)
	}
	if result.FulfillmentStatus != PhysicalFulfillmentStatusPendingDigitalConfirmation || result.FulfillmentStatusText == "" {
		t.Fatalf("expected pending digital status text for blocked asset, got %+v", result)
	}

	if err := db.Table(CardInstanceRow{}.TableName()).
		Where("id = 1").
		Updates(map[string]interface{}{
			"mint_status":  MintStatusSuccess,
			"chain_status": ChainStatusSuccess,
			"token_id":     "free-token-001",
		}).Error; err != nil {
		t.Fatalf("prepare success asset failed: %v", err)
	}
	if err := db.Exec(`UPDATE sms_draw_activity SET real_name_required = 1 WHERE id = 2001`).Error; err != nil {
		t.Fatalf("prepare activity real name failed: %v", err)
	}
	if err := db.Exec(`UPDATE ums_member_identity SET real_name_status = 'pending' WHERE member_id = 3001`).Error; err != nil {
		t.Fatalf("prepare member identity failed: %v", err)
	}
	result, err = service.EnsurePhysicalFulfillmentByAsset(context.Background(), physicalTestScope(), PhysicalFulfillmentInput{AssetInstanceID: 1})
	if err != nil {
		t.Fatalf("EnsurePhysicalFulfillmentByAsset real name returned error: %v", err)
	}
	if result.BlockedReason != PhysicalBlockRealName {
		t.Fatalf("expected real name block, got %+v", result)
	}
	if result.FulfillmentStatus != PhysicalFulfillmentStatusPendingRealName || result.FulfillmentStatusText == "" {
		t.Fatalf("expected pending real name status text for blocked asset, got %+v", result)
	}

	var count int64
	if err := db.Table(PhysicalFulfillmentRow{}.TableName()).Count(&count).Error; err != nil {
		t.Fatalf("count fulfillment failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no fulfillment row for blocked assets, got %d", count)
	}
}

func TestConfirmAddressSnapshotsAndMasksMemberDetail(t *testing.T) {
	db := preparePhysicalFulfillmentTestDB(t)
	service := NewService(db, nil, nil)

	created, err := service.EnsurePhysicalFulfillmentByAsset(context.Background(), physicalTestScope(), PhysicalFulfillmentInput{AssetInstanceID: 1})
	if err != nil {
		t.Fatalf("EnsurePhysicalFulfillmentByAsset returned error: %v", err)
	}
	confirmed, err := service.ConfirmPhysicalFulfillmentAddress(context.Background(), physicalTestScope(), ConfirmPhysicalFulfillmentAddressInput{
		FulfillmentID:   created.FulfillmentID,
		AssetInstanceID: 1,
		MemberID:        3001,
		AddressID:       501,
		RequestID:       "req-address",
		TraceID:         "trace-address",
	})
	if err != nil {
		t.Fatalf("ConfirmPhysicalFulfillmentAddress returned error: %v", err)
	}
	if confirmed.FulfillmentStatus != PhysicalFulfillmentStatusProductionPending {
		t.Fatalf("expected production pending after address confirmation, got %+v", confirmed)
	}

	detail, err := service.QueryMemberPhysicalFulfillmentDetail(context.Background(), physicalTestScope(), 3001, 1)
	if err != nil {
		t.Fatalf("QueryMemberPhysicalFulfillmentDetail returned error: %v", err)
	}
	if detail.ReceiverNameMasked != "张*" || detail.ReceiverPhoneMasked != "138****8001" {
		t.Fatalf("expected masked address, got %+v", detail)
	}
	if strings.Contains(fmt.Sprintf("%+v", detail), "13800138001") {
		t.Fatalf("member detail leaked full receiver phone: %+v", detail)
	}
	if len(detail.Timeline) < 2 {
		t.Fatalf("expected creation and address logs, got %+v", detail.Timeline)
	}
}

func TestConfirmPhysicalFulfillmentShippingFeeMarksMemberDetailPaid(t *testing.T) {
	db := preparePhysicalFulfillmentTestDB(t)
	service := NewService(db, nil, nil)

	created, err := service.EnsurePhysicalFulfillmentByAsset(context.Background(), physicalTestScope(), PhysicalFulfillmentInput{AssetInstanceID: 1})
	if err != nil {
		t.Fatalf("EnsurePhysicalFulfillmentByAsset returned error: %v", err)
	}
	_, err = service.ConfirmPhysicalFulfillmentShippingFee(context.Background(), physicalTestScope(), ConfirmPhysicalFulfillmentShippingFeeInput{
		FulfillmentID:   created.FulfillmentID,
		AssetInstanceID: 1,
		MemberID:        3001,
		PayAmount:       1200,
		PayChannel:      "test",
		PaymentNo:       "PAY-FEE-1",
	})
	if err != nil {
		t.Fatalf("ConfirmPhysicalFulfillmentShippingFee returned error: %v", err)
	}

	detail, err := service.QueryMemberPhysicalFulfillmentDetail(context.Background(), physicalTestScope(), 3001, 1)
	if err != nil {
		t.Fatalf("QueryMemberPhysicalFulfillmentDetail returned error: %v", err)
	}
	if detail.ShippingFeeStatus != PhysicalShippingFeeStatusPaid || detail.ShippingFeeAmount != 1200 {
		t.Fatalf("expected paid shipping fee detail, got %+v", detail)
	}
	hasFeeAction := false
	for _, item := range detail.Timeline {
		if item.Action == PhysicalActionShippingFeePaid {
			hasFeeAction = true
			break
		}
	}
	if !hasFeeAction {
		t.Fatalf("expected shipping fee action in timeline, got %+v", detail.Timeline)
	}
}

func TestPhysicalFulfillmentProductionShipSignExceptionAndReissue(t *testing.T) {
	db := preparePhysicalFulfillmentTestDB(t)
	service := NewService(db, nil, nil)

	created, err := service.EnsurePhysicalFulfillmentByAsset(context.Background(), physicalTestScope(), PhysicalFulfillmentInput{AssetInstanceID: 1, AddressID: 501})
	if err != nil {
		t.Fatalf("EnsurePhysicalFulfillmentByAsset returned error: %v", err)
	}
	if created.FulfillmentStatus != PhysicalFulfillmentStatusProductionPending {
		t.Fatalf("expected production pending with address, got %+v", created)
	}

	if _, err = service.UpdatePhysicalCardProductionStatus(context.Background(), physicalTestScope(), UpdatePhysicalCardProductionStatusInput{
		FulfillmentID:     created.FulfillmentID,
		OperatorID:        9001,
		ProductionBatchNo: "BATCH-001",
		ProductionStatus:  PhysicalProductionStatusPrinting,
		Reason:            "开始制作",
	}); err != nil {
		t.Fatalf("UpdatePhysicalCardProductionStatus returned error: %v", err)
	}
	if _, err = service.UpdatePhysicalCardProductionStatus(context.Background(), physicalTestScope(), UpdatePhysicalCardProductionStatusInput{
		FulfillmentID:     created.FulfillmentID,
		OperatorID:        9001,
		ProductionBatchNo: "BATCH-001",
		ProductionStatus:  PhysicalProductionStatusCompleted,
		Reason:            "制作完成",
	}); err != nil {
		t.Fatalf("complete production returned error: %v", err)
	}
	shipped, err := service.ShipPhysicalCard(context.Background(), physicalTestScope(), ShipPhysicalCardInput{
		FulfillmentID: created.FulfillmentID,
		OperatorID:    9001,
		CarrierCode:   "sf",
		CarrierName:   "顺丰速运",
		TrackingNo:    "SF123456789",
		Reason:        "发货",
	})
	if err != nil {
		t.Fatalf("ShipPhysicalCard returned error: %v", err)
	}
	if shipped.FulfillmentStatus != PhysicalFulfillmentStatusShipped || shipped.ShippingStatus != PhysicalShippingStatusShipped {
		t.Fatalf("expected shipped status, got %+v", shipped)
	}
	signed, err := service.ConfirmPhysicalCardReceipt(context.Background(), physicalTestScope(), ConfirmPhysicalCardReceiptInput{
		FulfillmentID: created.FulfillmentID,
		MemberID:      3001,
		Reason:        "本人签收",
	})
	if err != nil {
		t.Fatalf("ConfirmPhysicalCardReceipt returned error: %v", err)
	}
	if signed.FulfillmentStatus != PhysicalFulfillmentStatusSigned {
		t.Fatalf("expected signed status, got %+v", signed)
	}
	exceptionResult, err := service.MarkPhysicalFulfillmentException(context.Background(), physicalTestScope(), PhysicalFulfillmentExceptionInput{
		FulfillmentID: created.FulfillmentID,
		OperatorID:    9001,
		FailureCode:   "package_damaged",
		Reason:        "签收后反馈卡面破损",
	})
	if err != nil {
		t.Fatalf("MarkPhysicalFulfillmentException returned error: %v", err)
	}
	if exceptionResult.FulfillmentStatus != PhysicalFulfillmentStatusException {
		t.Fatalf("expected exception status, got %+v", exceptionResult)
	}
	reissue, err := service.RequestPhysicalCardReissue(context.Background(), physicalTestScope(), PhysicalFulfillmentExceptionInput{
		FulfillmentID: created.FulfillmentID,
		OperatorID:    9001,
		FailureCode:   "package_damaged",
		Reason:        "安排补发",
	})
	if err != nil {
		t.Fatalf("RequestPhysicalCardReissue returned error: %v", err)
	}
	if reissue.FulfillmentStatus != PhysicalFulfillmentStatusReissuePending {
		t.Fatalf("expected reissue pending status, got %+v", reissue)
	}

	var logCount int64
	if err := db.Table(PhysicalFulfillmentLogRow{}.TableName()).Where("fulfillment_id = ?", created.FulfillmentID).Count(&logCount).Error; err != nil {
		t.Fatalf("count logs failed: %v", err)
	}
	if logCount < 7 {
		t.Fatalf("expected fulfillment action logs, got %d", logCount)
	}
}
