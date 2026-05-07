package digital_card_chain

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
	"github.com/feihua/zero-admin/rpc/sys/client/operatelogservice"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"google.golang.org/grpc"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newDigitalCardChainAdminTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	stmts := []string{
		`CREATE TABLE sms_draw_participation_record (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			activity_id INTEGER NOT NULL,
			member_id INTEGER NOT NULL,
			pool_id INTEGER NOT NULL DEFAULT 0,
			template_id INTEGER NOT NULL DEFAULT 0,
			request_id TEXT NOT NULL DEFAULT '',
			trace_id TEXT NOT NULL DEFAULT '',
			eligibility_snapshot_json TEXT NOT NULL DEFAULT '',
			asset_instance_id INTEGER NOT NULL DEFAULT 0,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE sms_draw_activity (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL DEFAULT '',
			platform_id INTEGER NOT NULL DEFAULT 1,
			tenant_id INTEGER NOT NULL DEFAULT 0,
			merchant_id INTEGER NOT NULL DEFAULT 0,
			real_name_required INTEGER NOT NULL DEFAULT 0,
			publish_readiness INTEGER NOT NULL DEFAULT 1,
			copyright_status INTEGER NOT NULL DEFAULT 2,
			content_audit_status INTEGER NOT NULL DEFAULT 2,
			status INTEGER NOT NULL DEFAULT 1,
			audit_status INTEGER NOT NULL DEFAULT 2,
			is_enabled INTEGER NOT NULL DEFAULT 1,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE sms_card_template (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			template_name TEXT NOT NULL DEFAULT '',
			display_status INTEGER NOT NULL DEFAULT 1,
			content_audit_status INTEGER NOT NULL DEFAULT 2,
			credential_ref TEXT NOT NULL DEFAULT '',
			status INTEGER NOT NULL DEFAULT 1,
			audit_status INTEGER NOT NULL DEFAULT 2,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE sms_card_instance (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			platform_id INTEGER NOT NULL DEFAULT 1,
			tenant_id INTEGER NOT NULL DEFAULT 0,
			merchant_id INTEGER NOT NULL DEFAULT 0,
			activity_id INTEGER NOT NULL,
			member_id INTEGER NOT NULL,
			participation_record_id INTEGER NOT NULL,
			request_id TEXT NOT NULL DEFAULT '',
			trace_id TEXT NOT NULL DEFAULT '',
			scope TEXT NOT NULL DEFAULT '',
			pool_id INTEGER NOT NULL DEFAULT 0,
			template_id INTEGER NOT NULL DEFAULT 0,
			rarity TEXT NOT NULL DEFAULT '',
			asset_no TEXT NOT NULL,
			asset_status TEXT NOT NULL DEFAULT '',
			mint_status TEXT NOT NULL DEFAULT '',
			token_id TEXT NOT NULL DEFAULT '',
			chain_status TEXT NOT NULL DEFAULT '',
			source_type TEXT NOT NULL DEFAULT 'draw',
			source_id INTEGER NOT NULL DEFAULT 0,
			fulfillment_rule_id INTEGER NOT NULL DEFAULT 0,
			transferable INTEGER NOT NULL DEFAULT 0,
			transfer_limit INTEGER NOT NULL DEFAULT 0,
			claim_condition TEXT NOT NULL DEFAULT '',
			redemption_condition TEXT NOT NULL DEFAULT '',
			refund_policy TEXT NOT NULL DEFAULT '',
			display_status TEXT NOT NULL DEFAULT '',
			compliance_status TEXT NOT NULL DEFAULT '',
			display_reason TEXT NOT NULL DEFAULT '',
			compliance_reason TEXT NOT NULL DEFAULT '',
			rule_snapshot_json TEXT NOT NULL DEFAULT '',
			last_receipt_at DATETIME NULL,
			mint_task_id INTEGER NOT NULL DEFAULT 0,
			disposed_at DATETIME NULL,
			disposed_by INTEGER NOT NULL DEFAULT 0,
			issued_at DATETIME NULL,
			create_by INTEGER NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_by INTEGER NULL,
			update_time DATETIME NULL,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE sms_card_mint_task (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			platform_id INTEGER NOT NULL DEFAULT 1,
			tenant_id INTEGER NOT NULL DEFAULT 0,
			merchant_id INTEGER NOT NULL DEFAULT 0,
			asset_instance_id INTEGER NOT NULL,
			participation_record_id INTEGER NOT NULL DEFAULT 0,
			activity_id INTEGER NOT NULL DEFAULT 0,
			member_id INTEGER NOT NULL DEFAULT 0,
			request_id TEXT NOT NULL DEFAULT '',
			trace_id TEXT NOT NULL DEFAULT '',
			idempotency_key TEXT NOT NULL DEFAULT '',
			task_status TEXT NOT NULL DEFAULT '',
			mint_status TEXT NOT NULL DEFAULT '',
			chain_status TEXT NOT NULL DEFAULT '',
			token_id TEXT NOT NULL DEFAULT '',
			chain_tx_id TEXT NOT NULL DEFAULT '',
			retry_count INTEGER NOT NULL DEFAULT 0,
			max_retry_count INTEGER NOT NULL DEFAULT 3,
			last_error_code TEXT NOT NULL DEFAULT '',
			last_error_reason TEXT NOT NULL DEFAULT '',
			last_receipt_summary TEXT NOT NULL DEFAULT '',
			last_receipt_json TEXT NOT NULL DEFAULT '',
			last_execute_at DATETIME NULL,
			next_retry_at DATETIME NULL,
			manual_required INTEGER NOT NULL DEFAULT 0,
			frozen INTEGER NOT NULL DEFAULT 0,
			freeze_reason TEXT NOT NULL DEFAULT '',
			create_by INTEGER NOT NULL DEFAULT 0,
			update_by INTEGER NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NULL,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE UNIQUE INDEX uk_card_mint_task_asset ON sms_card_mint_task(asset_instance_id, is_deleted)`,
		`CREATE UNIQUE INDEX uk_card_mint_task_idempotency ON sms_card_mint_task(idempotency_key, is_deleted)`,
		`CREATE TABLE sms_card_asset_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			asset_instance_id INTEGER NOT NULL,
			participation_record_id INTEGER NOT NULL,
			from_status TEXT NOT NULL DEFAULT '',
			to_status TEXT NOT NULL DEFAULT '',
			operation_type TEXT NOT NULL DEFAULT '',
			operator_type TEXT NOT NULL DEFAULT '',
			trace_id TEXT NOT NULL DEFAULT '',
			reason_code TEXT NOT NULL DEFAULT '',
			reason_text TEXT NOT NULL DEFAULT '',
			payload_json TEXT NOT NULL DEFAULT '',
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("exec schema failed: %v", err)
		}
	}

	if err := db.Exec(`
		INSERT INTO sms_draw_activity
			(id, name, platform_id, tenant_id, merchant_id, real_name_required, publish_readiness, copyright_status, content_audit_status, status, audit_status, is_enabled, is_deleted)
		VALUES
			(2001, '春季抽卡', 1, 10, 88, 0, 1, 2, 2, 1, 2, 1, 0)
	`).Error; err != nil {
		t.Fatalf("seed activity failed: %v", err)
	}
	if err := db.Exec(`
		INSERT INTO sms_card_template
			(id, template_name, display_status, content_audit_status, credential_ref, status, audit_status, is_deleted)
		VALUES
			(21, 'SSR 兔兔', 1, 2, 'cred-1', 1, 2, 0)
	`).Error; err != nil {
		t.Fatalf("seed template failed: %v", err)
	}
	if err := db.Exec(`
		INSERT INTO sms_draw_participation_record
			(id, activity_id, member_id, pool_id, template_id, request_id, trace_id, eligibility_snapshot_json, asset_instance_id, is_deleted)
		VALUES
			(1, 2001, 3001, 11, 21, 'req-admin', 'trace-admin', '{"realNameStatus":"verified"}', 1, 0)
	`).Error; err != nil {
		t.Fatalf("seed participation record failed: %v", err)
	}
	if err := db.Exec(`
		INSERT INTO sms_card_instance
			(id, platform_id, tenant_id, merchant_id, activity_id, member_id, participation_record_id, request_id, trace_id, scope, pool_id, template_id, rarity, asset_no, asset_status, mint_status, token_id, chain_status, mint_task_id, create_by, is_deleted)
		VALUES
			(1, 1, 10, 88, 2001, 3001, 1, 'req-admin', 'trace-admin', 'platform:1,tenant:10,merchant:88', 11, 21, 'SSR', 'CARD-ADMIN', 'asset_created', 'mint_pending', '', '', 0, 0, 0)
	`).Error; err != nil {
		t.Fatalf("seed card instance failed: %v", err)
	}

	return db
}

func newAdminCtx() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "scopeType", "merchant")
	ctx = context.WithValue(ctx, "platformId", json.Number("1"))
	ctx = context.WithValue(ctx, "tenantId", json.Number("10"))
	ctx = context.WithValue(ctx, "merchantId", json.Number("88"))
	ctx = context.WithValue(ctx, "userId", json.Number("9001"))
	return ctx
}

type localChainAdminService struct {
	service *digitalcardmint.Service
}

func (s *localChainAdminService) QueryTaskList(ctx context.Context, scope pkgscope.GovernanceScope, filter digitalcardmint.QueryFilter, _ ...grpc.CallOption) (int64, []*digitalcardmint.TaskListItem, error) {
	return s.service.QueryTaskList(ctx, scope, filter)
}

func (s *localChainAdminService) QueryTaskDetail(ctx context.Context, scope pkgscope.GovernanceScope, taskID int64, _ ...grpc.CallOption) (*digitalcardmint.TaskDetail, error) {
	return s.service.QueryTaskDetail(ctx, scope, taskID)
}

func (s *localChainAdminService) QueryAvailableTaskActions(ctx context.Context, scope pkgscope.GovernanceScope, taskID int64, _ ...grpc.CallOption) ([]string, error) {
	return s.service.QueryAvailableActions(ctx, scope, taskID)
}

func (s *localChainAdminService) RetryTask(ctx context.Context, scope pkgscope.GovernanceScope, taskID int64, operatorID int64, reason string, _ ...grpc.CallOption) (*digitalcardmint.ActionResult, error) {
	return s.service.RetryTask(ctx, scope, taskID, operatorID, reason)
}

func (s *localChainAdminService) FreezeTask(ctx context.Context, scope pkgscope.GovernanceScope, taskID int64, operatorID int64, reason string, _ ...grpc.CallOption) (*digitalcardmint.ActionResult, error) {
	return s.service.FreezeTask(ctx, scope, taskID, operatorID, reason)
}

func (s *localChainAdminService) EscalateTask(ctx context.Context, scope pkgscope.GovernanceScope, taskID int64, operatorID int64, reason string, _ ...grpc.CallOption) (*digitalcardmint.ActionResult, error) {
	return s.service.EscalateTask(ctx, scope, taskID, operatorID, reason)
}

func (s *localChainAdminService) QueryAssetAuditList(context.Context, pkgscope.GovernanceScope, digitalcardmint.DigitalCardAssetAuditFilter, ...grpc.CallOption) (int64, []digitalcardmint.DigitalCardAssetAuditItem, error) {
	return 0, nil, nil
}

func (s *localChainAdminService) QueryAssetAuditDetail(context.Context, pkgscope.GovernanceScope, int64, ...grpc.CallOption) (*digitalcardmint.DigitalCardAssetAuditDetail, error) {
	return nil, nil
}

func (s *localChainAdminService) ReviewAssetCompliance(context.Context, pkgscope.GovernanceScope, int64, int64, string, ...grpc.CallOption) (*digitalcardmint.DigitalCardAssetActionResult, error) {
	return nil, nil
}

func (s *localChainAdminService) OfflineAssetDisplay(context.Context, pkgscope.GovernanceScope, int64, int64, string, ...grpc.CallOption) (*digitalcardmint.DigitalCardAssetActionResult, error) {
	return nil, nil
}

func (s *localChainAdminService) RecycleAsset(context.Context, pkgscope.GovernanceScope, int64, int64, string, ...grpc.CallOption) (*digitalcardmint.DigitalCardAssetActionResult, error) {
	return nil, nil
}

func (s *localChainAdminService) QueryPhysicalFulfillmentList(context.Context, pkgscope.GovernanceScope, digitalcardmint.PhysicalFulfillmentFilter, ...grpc.CallOption) (int64, []digitalcardmint.PhysicalFulfillmentItem, error) {
	return 0, nil, nil
}

func (s *localChainAdminService) QueryPhysicalFulfillmentDetail(context.Context, pkgscope.GovernanceScope, int64, ...grpc.CallOption) (*digitalcardmint.PhysicalFulfillmentAdminDetail, error) {
	return nil, nil
}

func (s *localChainAdminService) EnsurePhysicalFulfillment(context.Context, pkgscope.GovernanceScope, digitalcardmint.PhysicalFulfillmentInput, ...grpc.CallOption) (*digitalcardmint.PhysicalFulfillmentResult, error) {
	return nil, nil
}

func (s *localChainAdminService) UpdatePhysicalCardProductionStatus(context.Context, pkgscope.GovernanceScope, digitalcardmint.UpdatePhysicalCardProductionStatusInput, ...grpc.CallOption) (*digitalcardmint.PhysicalFulfillmentResult, error) {
	return nil, nil
}

func (s *localChainAdminService) ShipPhysicalCard(context.Context, pkgscope.GovernanceScope, digitalcardmint.ShipPhysicalCardInput, ...grpc.CallOption) (*digitalcardmint.PhysicalFulfillmentResult, error) {
	return nil, nil
}

func (s *localChainAdminService) MarkPhysicalFulfillmentException(context.Context, pkgscope.GovernanceScope, digitalcardmint.PhysicalFulfillmentExceptionInput, ...grpc.CallOption) (*digitalcardmint.PhysicalFulfillmentResult, error) {
	return nil, nil
}

func (s *localChainAdminService) RequestPhysicalCardReissue(context.Context, pkgscope.GovernanceScope, digitalcardmint.PhysicalFulfillmentExceptionInput, ...grpc.CallOption) (*digitalcardmint.PhysicalFulfillmentResult, error) {
	return nil, nil
}

func newAdminServiceContext(t *testing.T) (*svc.ServiceContext, int64) {
	t.Helper()

	db := newDigitalCardChainAdminTestDB(t)
	cardMintService := digitalcardmint.NewService(db, nil, nil)

	var taskID int64
	if err := db.Transaction(func(tx *gorm.DB) error {
		task, err := cardMintService.EnsureTaskTx(context.Background(), tx, 1, digitalcardmint.OperatorSystem)
		if err != nil {
			return err
		}
		taskID = task.ID
		return nil
	}); err != nil {
		t.Fatalf("EnsureTaskTx returned error: %v", err)
	}

	return &svc.ServiceContext{CardMintAdminService: &localChainAdminService{service: cardMintService}}, taskID
}

func TestQueryDigitalCardChainListMapsFields(t *testing.T) {
	svcCtx, _ := newAdminServiceContext(t)
	logic := NewQueryDigitalCardChainListLogic(newAdminCtx(), svcCtx)

	resp, err := logic.QueryDigitalCardChainList(&types.QueryDigitalCardChainListReq{
		Current:      1,
		PageSize:     10,
		ActivityName: "春季",
	})
	if err != nil {
		t.Fatalf("QueryDigitalCardChainList returned error: %v", err)
	}
	if resp.Total != 1 || len(resp.Data.List) != 1 {
		t.Fatalf("expected one mapped item, got total=%d len=%d", resp.Total, len(resp.Data.List))
	}

	item := resp.Data.List[0]
	if item.AssetNo != "CARD-ADMIN" || item.ActivityName != "春季抽卡" || item.TemplateName != "SSR 兔兔" {
		t.Fatalf("unexpected mapped item: %+v", item)
	}
	if item.AssetStatusText == "" {
		t.Fatalf("expected assetStatusText to be populated, got %+v", item)
	}
}

func TestQueryDigitalCardChainActionsBoundaries(t *testing.T) {
	svcCtx, taskID := newAdminServiceContext(t)
	logic := NewQueryDigitalCardChainActionsLogic(newAdminCtx(), svcCtx)

	resp, err := logic.QueryDigitalCardChainActions(&types.QueryDigitalCardChainActionsReq{TaskId: taskID})
	if err != nil {
		t.Fatalf("QueryDigitalCardChainActions returned error: %v", err)
	}
	actions := strings.Join(resp.AvailableActions, ",")
	for _, action := range []string{"retry", "freeze", "escalate"} {
		if !strings.Contains(actions, action) {
			t.Fatalf("expected action %s in %v", action, resp.AvailableActions)
		}
	}
}

func TestRetryDigitalCardChainRejectsCrossScope(t *testing.T) {
	svcCtx, taskID := newAdminServiceContext(t)
	logic := NewRetryDigitalCardChainLogic(newAdminCtx(), svcCtx)

	_, err := logic.RetryDigitalCardChain(&types.RetryDigitalCardChainReq{
		TaskId:     taskID,
		Reason:     "越权测试",
		ScopeType:  "merchant",
		PlatformId: 1,
		TenantId:   10,
		MerchantId: 99,
	})
	if err == nil {
		t.Fatal("expected cross-scope retry to fail")
	}
	if !strings.Contains(err.Error(), "当前主体不允许切换写入范围") {
		t.Fatalf("unexpected error: %v", err)
	}
}

type captureOperateLogService struct {
	req *sysclient.AddOperateLogReq
}

func (c *captureOperateLogService) AddOperateLog(_ context.Context, in *operatelogservice.AddOperateLogReq, _ ...grpc.CallOption) (*operatelogservice.AddOperateLogResp, error) {
	if in != nil {
		c.req = &sysclient.AddOperateLogReq{}
		*c.req = *in
	}
	return &operatelogservice.AddOperateLogResp{}, nil
}

func (c *captureOperateLogService) DeleteOperateLog(context.Context, *operatelogservice.DeleteOperateLogReq, ...grpc.CallOption) (*operatelogservice.DeleteOperateLogResp, error) {
	return &operatelogservice.DeleteOperateLogResp{}, nil
}

func (c *captureOperateLogService) QueryOperateLogDetail(context.Context, *operatelogservice.QueryOperateLogDetailReq, ...grpc.CallOption) (*operatelogservice.QueryOperateLogDetailResp, error) {
	return &operatelogservice.QueryOperateLogDetailResp{}, nil
}

func (c *captureOperateLogService) QueryOperateLogList(context.Context, *operatelogservice.QueryOperateLogListReq, ...grpc.CallOption) (*operatelogservice.QueryOperateLogListResp, error) {
	return &operatelogservice.QueryOperateLogListResp{}, nil
}

func TestWriteDigitalCardChainOperateLogEscapesReasonJSON(t *testing.T) {
	logger := &captureOperateLogService{}
	svcCtx := &svc.ServiceContext{Operatelogservice: logger}
	reason := "人工\"重试\"\n第二行\\说明"

	writeDigitalCardChainOperateLog(newAdminCtx(), svcCtx, 9001, "retry", 88, reason, &digitalcardmint.ActionResult{
		TaskStatus:  digitalcardmint.TaskStatusDispatched,
		MintStatus:  digitalcardmint.MintStatusProcessing,
		ChainStatus: digitalcardmint.ChainStatusProcessing,
	})

	if logger.req == nil {
		t.Fatal("expected operate log request to be captured")
	}

	var operateParam map[string]interface{}
	if err := json.Unmarshal([]byte(logger.req.OperateParam), &operateParam); err != nil {
		t.Fatalf("operate param should be valid json: %v", err)
	}
	if got := operateParam["reason"]; got != reason {
		t.Fatalf("expected reason to round-trip, got %#v", got)
	}

	var jsonResult map[string]interface{}
	if err := json.Unmarshal([]byte(logger.req.JsonResult), &jsonResult); err != nil {
		t.Fatalf("json result should be valid json: %v", err)
	}
	if jsonResult["taskStatus"] != digitalcardmint.TaskStatusDispatched {
		t.Fatalf("expected task status to be serialized, got %#v", jsonResult["taskStatus"])
	}

	var extra map[string]interface{}
	if err := json.Unmarshal([]byte(logger.req.Extra), &extra); err != nil {
		t.Fatalf("extra should be valid json: %v", err)
	}
	if extra["action"] != "retry" {
		t.Fatalf("expected action to be serialized, got %#v", extra["action"])
	}
}
