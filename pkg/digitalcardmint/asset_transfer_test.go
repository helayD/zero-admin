package digitalcardmint

import (
	"context"
	"strings"
	"testing"
)

func prepareAssetTransferTestDB(t *testing.T) *Service {
	t.Helper()
	db := newAssetServiceTestDB(t)
	stmts := []string{
		`CREATE TABLE ums_member_info (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			member_id INTEGER NOT NULL,
			nickname TEXT NOT NULL DEFAULT '',
			mobile TEXT NOT NULL DEFAULT '',
			is_enabled INTEGER NOT NULL DEFAULT 1,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE sms_card_physical_fulfillment (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			asset_instance_id INTEGER NOT NULL,
			member_id INTEGER NOT NULL DEFAULT 0,
			fulfillment_status TEXT NOT NULL DEFAULT '',
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
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
		`INSERT INTO ums_member_info (member_id, nickname, mobile, is_enabled, is_deleted) VALUES
			(3002, '接收人', '13800138002', 1, 0),
			(3999, '赠送人', '13800138999', 1, 0)`,
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("prepare transfer schema failed: %v", err)
		}
	}
	return NewService(db, nil, nil)
}

func TestResolveAssetTransferRecipientReturnsRegisterH5ForUnknownMobile(t *testing.T) {
	service := prepareAssetTransferTestDB(t)

	result, err := service.ResolveAssetTransferRecipient(context.Background(), merchantScope(), AssetTransferRecipientInput{
		AssetInstanceID:   3,
		FromMemberID:      3999,
		RecipientMobile:   "13800138009",
		RegisterH5BaseURL: "/h5/digitalCard/register",
	})
	if err != nil {
		t.Fatalf("ResolveAssetTransferRecipient returned error: %v", err)
	}
	if result.RecipientStatus != TransferRecipientStatusNeedRegister || result.CanTransfer {
		t.Fatalf("expected need register response, got %+v", result)
	}
	if !strings.Contains(result.RegisterURL, "/h5/digitalCard/register") || !strings.Contains(result.RegisterURL, "assetInstanceId=3") {
		t.Fatalf("expected h5 register url, got %s", result.RegisterURL)
	}
}

func TestTransferDigitalCardAssetMovesOwnershipToRegisteredMember(t *testing.T) {
	service := prepareAssetTransferTestDB(t)

	result, err := service.TransferDigitalCardAsset(context.Background(), merchantScope(), TransferDigitalCardAssetInput{
		AssetInstanceID: 3,
		FromMemberID:    3999,
		RecipientMobile: "13800138002",
		RequestID:       "transfer-req-1",
	})
	if err != nil {
		t.Fatalf("TransferDigitalCardAsset returned error: %v", err)
	}
	if result.ToMemberID != 3002 || result.TransferStatus != "transferred" {
		t.Fatalf("unexpected transfer result: %+v", result)
	}

	total, list, err := service.QueryMemberDigitalCardAssetList(context.Background(), merchantScope(), 3002, MemberDigitalCardAssetFilter{PageNum: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("QueryMemberDigitalCardAssetList returned error: %v", err)
	}
	if total != 1 || len(list) != 1 || list[0].AssetNo != "CARD-003" {
		t.Fatalf("expected transferred asset for recipient, total=%d list=%+v", total, list)
	}

	logs, err := service.queryAssetLogs(context.Background(), 3)
	if err != nil {
		t.Fatalf("query logs failed: %v", err)
	}
	if len(logs) == 0 || logs[0].OperationType != OperationAssetTransferred {
		t.Fatalf("expected transfer log first, got %+v", logs)
	}
}

func TestRequestDigitalCardAssetWithdrawAppendsAuditLog(t *testing.T) {
	service := prepareAssetTransferTestDB(t)

	result, err := service.RequestDigitalCardAssetWithdraw(context.Background(), merchantScope(), RequestDigitalCardAssetWithdrawInput{
		AssetInstanceID: 3,
		MemberID:        3999,
		Reason:          "用户申请提现",
	})
	if err != nil {
		t.Fatalf("RequestDigitalCardAssetWithdraw returned error: %v", err)
	}
	if result.WithdrawStatus != "requested" {
		t.Fatalf("expected requested status, got %+v", result)
	}
	logs, err := service.queryAssetLogs(context.Background(), 3)
	if err != nil {
		t.Fatalf("query logs failed: %v", err)
	}
	if len(logs) == 0 || logs[0].OperationType != OperationAssetWithdrawRequested {
		t.Fatalf("expected withdraw log first, got %+v", logs)
	}
}
