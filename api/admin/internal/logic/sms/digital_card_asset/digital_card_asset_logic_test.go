package digital_card_asset

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"google.golang.org/grpc"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newAdminAssetTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	stmts := []string{
		`CREATE TABLE sms_draw_activity (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL DEFAULT '', compliance_rule_summary TEXT NOT NULL DEFAULT '', platform_id INTEGER NOT NULL DEFAULT 1, tenant_id INTEGER NOT NULL DEFAULT 0, merchant_id INTEGER NOT NULL DEFAULT 0, is_deleted INTEGER NOT NULL DEFAULT 0)`,
		`CREATE TABLE sms_card_template (id INTEGER PRIMARY KEY AUTOINCREMENT, template_name TEXT NOT NULL DEFAULT '', card_face_image TEXT NOT NULL DEFAULT '', is_deleted INTEGER NOT NULL DEFAULT 0)`,
		`CREATE TABLE sms_draw_participation_record (id INTEGER PRIMARY KEY AUTOINCREMENT, activity_id INTEGER NOT NULL, member_id INTEGER NOT NULL, request_id TEXT NOT NULL DEFAULT '', trace_id TEXT NOT NULL DEFAULT '', result_type TEXT NOT NULL DEFAULT '', result_status TEXT NOT NULL DEFAULT '', failure_reason TEXT NOT NULL DEFAULT '', create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, is_deleted INTEGER NOT NULL DEFAULT 0)`,
		`CREATE TABLE sms_card_instance (id INTEGER PRIMARY KEY AUTOINCREMENT, platform_id INTEGER NOT NULL DEFAULT 1, tenant_id INTEGER NOT NULL DEFAULT 0, merchant_id INTEGER NOT NULL DEFAULT 0, activity_id INTEGER NOT NULL, member_id INTEGER NOT NULL, participation_record_id INTEGER NOT NULL, request_id TEXT NOT NULL DEFAULT '', trace_id TEXT NOT NULL DEFAULT '', template_id INTEGER NOT NULL DEFAULT 0, rarity TEXT NOT NULL DEFAULT '', asset_no TEXT NOT NULL DEFAULT '', asset_status TEXT NOT NULL DEFAULT '', mint_status TEXT NOT NULL DEFAULT '', token_id TEXT NOT NULL DEFAULT '', chain_status TEXT NOT NULL DEFAULT '', display_status TEXT NOT NULL DEFAULT '', compliance_status TEXT NOT NULL DEFAULT '', display_reason TEXT NOT NULL DEFAULT '', compliance_reason TEXT NOT NULL DEFAULT '', rule_snapshot_json TEXT NOT NULL DEFAULT '', mint_task_id INTEGER NOT NULL DEFAULT 0, issued_at DATETIME NULL, disposed_at DATETIME NULL, disposed_by INTEGER NOT NULL DEFAULT 0, update_by INTEGER NOT NULL DEFAULT 0, create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, update_time DATETIME NULL, is_deleted INTEGER NOT NULL DEFAULT 0)`,
		`CREATE TABLE sms_card_mint_task (id INTEGER PRIMARY KEY AUTOINCREMENT, platform_id INTEGER NOT NULL DEFAULT 1, tenant_id INTEGER NOT NULL DEFAULT 0, merchant_id INTEGER NOT NULL DEFAULT 0, asset_instance_id INTEGER NOT NULL, participation_record_id INTEGER NOT NULL DEFAULT 0, activity_id INTEGER NOT NULL DEFAULT 0, member_id INTEGER NOT NULL DEFAULT 0, request_id TEXT NOT NULL DEFAULT '', trace_id TEXT NOT NULL DEFAULT '', task_status TEXT NOT NULL DEFAULT '', mint_status TEXT NOT NULL DEFAULT '', chain_status TEXT NOT NULL DEFAULT '', token_id TEXT NOT NULL DEFAULT '', chain_tx_id TEXT NOT NULL DEFAULT '', last_receipt_summary TEXT NOT NULL DEFAULT '', last_receipt_json TEXT NOT NULL DEFAULT '', create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, update_time DATETIME NULL, is_deleted INTEGER NOT NULL DEFAULT 0)`,
		`CREATE TABLE sms_card_asset_log (id INTEGER PRIMARY KEY AUTOINCREMENT, asset_instance_id INTEGER NOT NULL, participation_record_id INTEGER NOT NULL DEFAULT 0, from_status TEXT NOT NULL DEFAULT '', to_status TEXT NOT NULL DEFAULT '', operation_type TEXT NOT NULL DEFAULT '', operator_type TEXT NOT NULL DEFAULT '', trace_id TEXT NOT NULL DEFAULT '', reason_code TEXT NOT NULL DEFAULT '', reason_text TEXT NOT NULL DEFAULT '', payload_json TEXT NOT NULL DEFAULT '', create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP)`,
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("exec schema failed: %v", err)
		}
	}

	seeds := []string{
		`INSERT INTO sms_draw_activity (id, name, compliance_rule_summary, platform_id, tenant_id, merchant_id, is_deleted) VALUES (2001, '春季抽卡', '默认禁止收益承诺', 1, 10, 88, 0)`,
		`INSERT INTO sms_card_template (id, template_name, card_face_image, is_deleted) VALUES (21, 'SSR 兔兔', 'https://img.example.com/ssr-rabbit.png', 0)`,
		`INSERT INTO sms_draw_participation_record (id, activity_id, member_id, request_id, trace_id, result_type, result_status, failure_reason, create_time, is_deleted) VALUES
			(1, 2001, 3001, 'req-admin-1', 'trace-admin-1', 'won', 'won_pending_asset', '', '2026-04-18 10:00:00', 0),
			(2, 2001, 3002, 'req-admin-2', 'trace-admin-2', 'won', 'won_pending_asset', '', '2026-04-18 10:05:00', 0)`,
		`INSERT INTO sms_card_instance (id, platform_id, tenant_id, merchant_id, activity_id, member_id, participation_record_id, request_id, trace_id, template_id, rarity, asset_no, asset_status, mint_status, token_id, chain_status, display_status, compliance_status, display_reason, compliance_reason, rule_snapshot_json, mint_task_id, issued_at, create_time, update_time, is_deleted) VALUES
			(1, 1, 10, 88, 2001, 3001, 1, 'req-admin-1', 'trace-admin-1', 21, 'SSR', 'CARD-ADMIN-001', 'asset_created', 'mint_processing', '', 'processing', 'display_visible', 'compliance_clear', '', '', '', 11, '2026-04-18 10:00:00', '2026-04-18 10:00:00', '2026-04-18 10:01:00', 0),
			(2, 1, 10, 99, 2001, 3002, 2, 'req-admin-2', 'trace-admin-2', 21, 'SSR', 'CARD-ADMIN-002', 'asset_created', 'mint_success', 'token-admin-2', 'success', 'display_hidden', 'compliance_review', '复核中', '复核中', '{\"scene\":\"review\"}', 12, '2026-04-18 10:05:00', '2026-04-18 10:05:00', '2026-04-18 10:06:00', 0)`,
		`INSERT INTO sms_card_mint_task (id, platform_id, tenant_id, merchant_id, asset_instance_id, participation_record_id, activity_id, member_id, request_id, trace_id, task_status, mint_status, chain_status, token_id, chain_tx_id, last_receipt_summary, last_receipt_json, create_time, update_time, is_deleted) VALUES
			(11, 1, 10, 88, 1, 1, 2001, 3001, 'req-admin-1', 'trace-admin-1', 'running', 'mint_processing', 'processing', '', '', '链上处理中', '{\"status\":\"processing\"}', '2026-04-18 10:00:00', '2026-04-18 10:01:00', 0),
			(12, 1, 10, 99, 2, 2, 2001, 3002, 'req-admin-2', 'trace-admin-2', 'succeeded', 'mint_success', 'success', 'token-admin-2', 'tx-admin-2', '链上成功', '{\"status\":\"success\"}', '2026-04-18 10:05:00', '2026-04-18 10:06:00', 0)`,
		`INSERT INTO sms_card_asset_log (id, asset_instance_id, participation_record_id, from_status, to_status, operation_type, operator_type, trace_id, reason_code, reason_text, payload_json, create_time) VALUES
			(1, 1, 1, '', 'mint_processing', 'mint_dispatching', 'system', 'trace-admin-1', '', '链上处理中', '{}', '2026-04-18 10:01:00'),
			(2, 2, 2, 'compliance_clear', 'compliance_review', 'asset_compliance_review', 'manual', 'trace-admin-2', '', '复核中', '{}', '2026-04-18 10:06:00')`,
	}
	for _, seed := range seeds {
		if err := db.Exec(seed).Error; err != nil {
			t.Fatalf("seed data failed: %v", err)
		}
	}

	return db
}

func newAdminAssetCtx(scopeType, platformID, tenantID, merchantID, userID string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "scopeType", scopeType)
	ctx = context.WithValue(ctx, "platformId", json.Number(platformID))
	ctx = context.WithValue(ctx, "tenantId", json.Number(tenantID))
	ctx = context.WithValue(ctx, "merchantId", json.Number(merchantID))
	ctx = context.WithValue(ctx, "userId", json.Number(userID))
	return ctx
}

type localAssetAdminService struct {
	service *digitalcardmint.Service
}

func (s *localAssetAdminService) QueryTaskList(context.Context, pkgscope.GovernanceScope, digitalcardmint.QueryFilter, ...grpc.CallOption) (int64, []*digitalcardmint.TaskListItem, error) {
	return 0, nil, nil
}

func (s *localAssetAdminService) QueryTaskDetail(context.Context, pkgscope.GovernanceScope, int64, ...grpc.CallOption) (*digitalcardmint.TaskDetail, error) {
	return nil, nil
}

func (s *localAssetAdminService) QueryAvailableTaskActions(context.Context, pkgscope.GovernanceScope, int64, ...grpc.CallOption) ([]string, error) {
	return nil, nil
}

func (s *localAssetAdminService) RetryTask(context.Context, pkgscope.GovernanceScope, int64, int64, string, ...grpc.CallOption) (*digitalcardmint.ActionResult, error) {
	return nil, nil
}

func (s *localAssetAdminService) FreezeTask(context.Context, pkgscope.GovernanceScope, int64, int64, string, ...grpc.CallOption) (*digitalcardmint.ActionResult, error) {
	return nil, nil
}

func (s *localAssetAdminService) EscalateTask(context.Context, pkgscope.GovernanceScope, int64, int64, string, ...grpc.CallOption) (*digitalcardmint.ActionResult, error) {
	return nil, nil
}

func (s *localAssetAdminService) QueryAssetAuditList(ctx context.Context, scope pkgscope.GovernanceScope, filter digitalcardmint.DigitalCardAssetAuditFilter, _ ...grpc.CallOption) (int64, []digitalcardmint.DigitalCardAssetAuditItem, error) {
	return s.service.QueryDigitalCardAssetAuditList(ctx, scope, filter)
}

func (s *localAssetAdminService) QueryAssetAuditDetail(ctx context.Context, scope pkgscope.GovernanceScope, assetInstanceID int64, _ ...grpc.CallOption) (*digitalcardmint.DigitalCardAssetAuditDetail, error) {
	return s.service.QueryDigitalCardAssetAuditDetail(ctx, scope, assetInstanceID)
}

func (s *localAssetAdminService) ReviewAssetCompliance(ctx context.Context, scope pkgscope.GovernanceScope, assetInstanceID int64, operatorID int64, reason string, _ ...grpc.CallOption) (*digitalcardmint.DigitalCardAssetActionResult, error) {
	return s.service.ReviewDigitalCardAssetCompliance(ctx, scope, assetInstanceID, operatorID, reason)
}

func (s *localAssetAdminService) OfflineAssetDisplay(ctx context.Context, scope pkgscope.GovernanceScope, assetInstanceID int64, operatorID int64, reason string, _ ...grpc.CallOption) (*digitalcardmint.DigitalCardAssetActionResult, error) {
	return s.service.OfflineDigitalCardAssetDisplay(ctx, scope, assetInstanceID, operatorID, reason)
}

func (s *localAssetAdminService) RecycleAsset(ctx context.Context, scope pkgscope.GovernanceScope, assetInstanceID int64, operatorID int64, reason string, _ ...grpc.CallOption) (*digitalcardmint.DigitalCardAssetActionResult, error) {
	return s.service.RecycleDigitalCardAsset(ctx, scope, assetInstanceID, operatorID, reason)
}

func (s *localAssetAdminService) QueryPhysicalFulfillmentList(context.Context, pkgscope.GovernanceScope, digitalcardmint.PhysicalFulfillmentFilter, ...grpc.CallOption) (int64, []digitalcardmint.PhysicalFulfillmentItem, error) {
	return 0, nil, nil
}

func (s *localAssetAdminService) QueryPhysicalFulfillmentDetail(context.Context, pkgscope.GovernanceScope, int64, ...grpc.CallOption) (*digitalcardmint.PhysicalFulfillmentAdminDetail, error) {
	return nil, nil
}

func (s *localAssetAdminService) EnsurePhysicalFulfillment(context.Context, pkgscope.GovernanceScope, digitalcardmint.PhysicalFulfillmentInput, ...grpc.CallOption) (*digitalcardmint.PhysicalFulfillmentResult, error) {
	return nil, nil
}

func (s *localAssetAdminService) UpdatePhysicalCardProductionStatus(context.Context, pkgscope.GovernanceScope, digitalcardmint.UpdatePhysicalCardProductionStatusInput, ...grpc.CallOption) (*digitalcardmint.PhysicalFulfillmentResult, error) {
	return nil, nil
}

func (s *localAssetAdminService) ShipPhysicalCard(context.Context, pkgscope.GovernanceScope, digitalcardmint.ShipPhysicalCardInput, ...grpc.CallOption) (*digitalcardmint.PhysicalFulfillmentResult, error) {
	return nil, nil
}

func (s *localAssetAdminService) MarkPhysicalFulfillmentException(context.Context, pkgscope.GovernanceScope, digitalcardmint.PhysicalFulfillmentExceptionInput, ...grpc.CallOption) (*digitalcardmint.PhysicalFulfillmentResult, error) {
	return nil, nil
}

func (s *localAssetAdminService) RequestPhysicalCardReissue(context.Context, pkgscope.GovernanceScope, digitalcardmint.PhysicalFulfillmentExceptionInput, ...grpc.CallOption) (*digitalcardmint.PhysicalFulfillmentResult, error) {
	return nil, nil
}

func newAdminAssetServiceContext(t *testing.T) *svc.ServiceContext {
	t.Helper()
	return &svc.ServiceContext{
		CardMintAdminService: &localAssetAdminService{service: digitalcardmint.NewService(newAdminAssetTestDB(t), nil, nil)},
	}
}

func TestQueryDigitalCardAssetListRespectsMerchantScope(t *testing.T) {
	logic := NewQueryDigitalCardAssetListLogic(newAdminAssetCtx("merchant", "1", "10", "88", "9001"), newAdminAssetServiceContext(t))

	resp, err := logic.QueryDigitalCardAssetList(&types.QueryDigitalCardAssetListReq{
		Current:  1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("QueryDigitalCardAssetList returned error: %v", err)
	}
	if resp.Total != 1 || len(resp.Data.List) != 1 {
		t.Fatalf("expected only merchant 88 assets, got total=%d len=%d", resp.Total, len(resp.Data.List))
	}
	if resp.Data.List[0].AssetNo != "CARD-ADMIN-001" {
		t.Fatalf("unexpected list item: %+v", resp.Data.List[0])
	}
}

func TestReviewDigitalCardAssetComplianceRejectsCrossScopeWrite(t *testing.T) {
	logic := NewReviewDigitalCardAssetComplianceLogic(newAdminAssetCtx("merchant", "1", "10", "88", "9001"), newAdminAssetServiceContext(t))

	_, err := logic.ReviewDigitalCardAssetCompliance(&types.DigitalCardAssetActionReq{
		DigitalCardAssetGovernanceScopeReq: types.DigitalCardAssetGovernanceScopeReq{
			ScopeType:  "merchant",
			PlatformId: 1,
			TenantId:   10,
			MerchantId: 99,
		},
		AssetInstanceId: 1,
		Reason:          "越权尝试",
	})
	if err == nil || !strings.Contains(err.Error(), "当前主体不允许切换写入范围") {
		t.Fatalf("expected cross-scope write rejection, got %v", err)
	}
}

func TestReviewThenRecycleDigitalCardAssetSuccess(t *testing.T) {
	svcCtx := newAdminAssetServiceContext(t)
	ctx := newAdminAssetCtx("merchant", "1", "10", "88", "9001")

	reviewLogic := NewReviewDigitalCardAssetComplianceLogic(ctx, svcCtx)
	reviewResp, err := reviewLogic.ReviewDigitalCardAssetCompliance(&types.DigitalCardAssetActionReq{
		AssetInstanceId: 1,
		Reason:          "投诉升级复核",
	})
	if err != nil {
		t.Fatalf("ReviewDigitalCardAssetCompliance returned error: %v", err)
	}
	if reviewResp.DisplayStatus != digitalcardmint.DisplayStatusHidden || reviewResp.ComplianceStatus != digitalcardmint.ComplianceStatusReview {
		t.Fatalf("unexpected review response: %+v", reviewResp)
	}

	recycleLogic := NewRecycleDigitalCardAssetLogic(ctx, svcCtx)
	recycleResp, err := recycleLogic.RecycleDigitalCardAsset(&types.DigitalCardAssetActionReq{
		AssetInstanceId: 1,
		Reason:          "复核后回收",
	})
	if err != nil {
		t.Fatalf("RecycleDigitalCardAsset returned error: %v", err)
	}
	if recycleResp.DisplayStatus != digitalcardmint.DisplayStatusRecycled || recycleResp.ComplianceStatus != digitalcardmint.ComplianceStatusRecycled {
		t.Fatalf("unexpected recycle response: %+v", recycleResp)
	}

	detailLogic := NewQueryDigitalCardAssetDetailLogic(ctx, svcCtx)
	detailResp, err := detailLogic.QueryDigitalCardAssetDetail(&types.QueryDigitalCardAssetDetailReq{AssetInstanceId: 1})
	if err != nil {
		t.Fatalf("QueryDigitalCardAssetDetail returned error: %v", err)
	}
	if detailResp.Data.Item.DisplayStatus != digitalcardmint.DisplayStatusRecycled {
		t.Fatalf("expected recycled detail, got %+v", detailResp.Data.Item)
	}
	if len(detailResp.Data.Logs) == 0 || detailResp.Data.Logs[0].OperationType != digitalcardmint.OperationAssetRecycled {
		t.Fatalf("expected latest log to be recycle action, got %+v", detailResp.Data.Logs)
	}
}
