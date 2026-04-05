package channelintegrationtemplateservicelogic

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/feihua/zero-admin/pkg/channeltemplate"
	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newChannelIntegrationTemplateTestSvc(t *testing.T) *svc.ServiceContext {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "channel-integration-template-test.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	stmt := `CREATE TABLE sys_channel_integration_template (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		template_code TEXT NOT NULL,
		template_name TEXT NOT NULL,
		template_type TEXT NOT NULL,
		target_code TEXT NOT NULL,
		scope_type TEXT NOT NULL,
		platform_id INTEGER NOT NULL DEFAULT 1,
		tenant_id INTEGER NOT NULL DEFAULT 0,
		merchant_id INTEGER NOT NULL DEFAULT 0,
		status TEXT NOT NULL,
		metadata_config TEXT NOT NULL DEFAULT '{}',
		secret_ref_config TEXT NOT NULL DEFAULT '{}',
		intent_contract_config TEXT NOT NULL DEFAULT '{}',
		impact_scope_config TEXT NOT NULL DEFAULT '{}',
		remark TEXT NOT NULL DEFAULT '',
		create_by TEXT NOT NULL DEFAULT '',
		create_time DATETIME NOT NULL,
		update_by TEXT NOT NULL DEFAULT '',
		update_time DATETIME NULL
	)`
	if err := db.Exec(stmt).Error; err != nil {
		t.Fatalf("create test schema failed: %v", err)
	}
	if err := db.Exec(`CREATE TABLE sys_tenant (
		id INTEGER PRIMARY KEY,
		tenant_code TEXT NOT NULL DEFAULT '',
		tenant_name TEXT NOT NULL DEFAULT '',
		status INTEGER NOT NULL DEFAULT 1,
		available_channels TEXT NOT NULL DEFAULT '[]'
	)`).Error; err != nil {
		t.Fatalf("create tenant schema failed: %v", err)
	}
	if err := db.Exec(`CREATE TABLE sys_merchant (
		id INTEGER PRIMARY KEY,
		tenant_id INTEGER NOT NULL DEFAULT 0,
		merchant_code TEXT NOT NULL DEFAULT '',
		merchant_name TEXT NOT NULL DEFAULT '',
		business_status INTEGER NOT NULL DEFAULT 1,
		available_channels TEXT NOT NULL DEFAULT '[]'
	)`).Error; err != nil {
		t.Fatalf("create merchant schema failed: %v", err)
	}
	if err := db.Exec(`CREATE TABLE sys_channel_integration_binding (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		template_id INTEGER NOT NULL,
		subject_type TEXT NOT NULL,
		platform_id INTEGER NOT NULL DEFAULT 1,
		tenant_id INTEGER NOT NULL DEFAULT 0,
		merchant_id INTEGER NOT NULL DEFAULT 0,
		target_code TEXT NOT NULL,
		binding_status TEXT NOT NULL,
		binding_source TEXT NOT NULL DEFAULT '',
		effect_scope_snapshot TEXT NOT NULL DEFAULT '{}',
		published_by TEXT NOT NULL DEFAULT '',
		published_at DATETIME NULL,
		remark TEXT NOT NULL DEFAULT '',
		create_by TEXT NOT NULL DEFAULT '',
		create_time DATETIME NOT NULL,
		update_by TEXT NOT NULL DEFAULT '',
		update_time DATETIME NULL
	)`).Error; err != nil {
		t.Fatalf("create binding schema failed: %v", err)
	}
	if err := db.Exec(`CREATE TABLE sys_operate_log (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL DEFAULT '',
		business_type INTEGER NOT NULL DEFAULT 0,
		method TEXT NOT NULL DEFAULT '',
		request_method TEXT NOT NULL DEFAULT '',
		operator_type INTEGER NOT NULL DEFAULT 0,
		operate_name TEXT NOT NULL DEFAULT '',
		dept_name TEXT NOT NULL DEFAULT '',
		operate_url TEXT NOT NULL DEFAULT '',
		operate_ip TEXT NOT NULL DEFAULT '',
		operate_location TEXT NOT NULL DEFAULT '',
		operate_param TEXT NOT NULL DEFAULT '',
		json_result TEXT NOT NULL DEFAULT '',
		platform TEXT NOT NULL DEFAULT '',
		browser TEXT NOT NULL DEFAULT '',
		version TEXT NOT NULL DEFAULT '',
		os TEXT NOT NULL DEFAULT '',
		arch TEXT NOT NULL DEFAULT '',
		engine TEXT NOT NULL DEFAULT '',
		engine_details TEXT NOT NULL DEFAULT '',
		extra TEXT NOT NULL DEFAULT '',
		status INTEGER NOT NULL DEFAULT 0,
		error_msg TEXT NOT NULL DEFAULT '',
		operate_time DATETIME NOT NULL,
		cost_time INTEGER NOT NULL DEFAULT 0
	)`).Error; err != nil {
		t.Fatalf("create operate log schema failed: %v", err)
	}
	if err := db.Exec(`INSERT INTO sys_tenant (id, tenant_code, tenant_name, status, available_channels) VALUES
		(1001, 'tenant_1001', '租户1001', 1, '["h5","mini_program"]')`).Error; err != nil {
		t.Fatalf("seed tenant failed: %v", err)
	}
	if err := db.Exec(`INSERT INTO sys_merchant (id, tenant_id, merchant_code, merchant_name, business_status, available_channels) VALUES
		(2001, 1001, 'merchant_2001', '商户2001', 1, '["mini_program"]')`).Error; err != nil {
		t.Fatalf("seed merchant failed: %v", err)
	}

	return &svc.ServiceContext{DB: db}
}

func TestCreateChannelIntegrationTemplateNormalizesAliasAndSecretRefs(t *testing.T) {
	svcCtx := newChannelIntegrationTemplateTestSvc(t)
	logic := NewCreateChannelIntegrationTemplateLogic(context.Background(), svcCtx)

	resp, err := logic.CreateChannelIntegrationTemplate(&sysclient.CreateChannelIntegrationTemplateReq{
		TemplateCode:         "mini_program_default",
		TemplateName:         "小程序模板",
		TemplateType:         channeltemplate.TemplateTypeChannel,
		TargetCode:           "mini-program",
		ScopeType:            "platform",
		Status:               channeltemplate.StatusEnabled,
		MetadataConfig:       `{"channelCode":"mini_program"}`,
		SecretRefConfig:      `{"appSecretRef":"credential://sys/channel/mini_program/app-secret"}`,
		IntentContractConfig: `{"home":{"intent":"home","routeKey":"home"}}`,
		ImpactScopeConfig:    `{"subjectTypes":["tenant","merchant"]}`,
		Remark:               "seed",
		CreateBy:             "tester",
	})
	if err != nil {
		t.Fatalf("CreateChannelIntegrationTemplate returned error: %v", err)
	}
	if resp.Id <= 0 {
		t.Fatalf("unexpected create response: %+v", resp)
	}

	var row channelIntegrationTemplateRow
	if err := svcCtx.DB.First(&row, resp.Id).Error; err != nil {
		t.Fatalf("load template failed: %v", err)
	}
	if row.TargetCode != channeltemplate.TargetMiniProgram {
		t.Fatalf("unexpected target code: %s", row.TargetCode)
	}
	if !strings.Contains(row.SecretRefConfig, "credential://") {
		t.Fatalf("unexpected secret_ref_config: %s", row.SecretRefConfig)
	}
}

func TestQueryChannelIntegrationTemplateListSupportsAliasFilter(t *testing.T) {
	svcCtx := newChannelIntegrationTemplateTestSvc(t)
	if err := svcCtx.DB.Exec(`INSERT INTO sys_channel_integration_template (
		template_code, template_name, template_type, target_code, scope_type, platform_id, tenant_id, merchant_id, status,
		metadata_config, secret_ref_config, intent_contract_config, impact_scope_config, remark, create_by, create_time, update_by
	) VALUES
		('tpl_h5', 'H5 模板', 'channel', 'h5', 'platform', 1, 0, 0, 'enabled', '{}', '{}', '{"home":{"intent":"home","routeKey":"home"}}', '{}', '', 'seed', CURRENT_TIMESTAMP, 'seed'),
		('tpl_mini', '小程序模板', 'channel', 'mini_program', 'platform', 1, 0, 0, 'enabled', '{}', '{"appSecretRef":"credential://sys/channel/mini_program/app-secret"}', '{"home":{"intent":"home","routeKey":"home"}}', '{}', '', 'seed', CURRENT_TIMESTAMP, 'seed')`).Error; err != nil {
		t.Fatalf("seed templates failed: %v", err)
	}

	logic := NewQueryChannelIntegrationTemplateListLogic(context.Background(), svcCtx)
	resp, err := logic.QueryChannelIntegrationTemplateList(&sysclient.QueryChannelIntegrationTemplateListReq{
		PageNum:      1,
		PageSize:     20,
		TemplateType: channeltemplate.TemplateTypeChannel,
		TargetCode:   "mini-program",
		Status:       channeltemplate.StatusEnabled,
	})
	if err != nil {
		t.Fatalf("QueryChannelIntegrationTemplateList returned error: %v", err)
	}
	if resp.Total != 1 || len(resp.List) != 1 {
		t.Fatalf("unexpected list response: %+v", resp)
	}
	if resp.List[0].TargetCode != channeltemplate.TargetMiniProgram {
		t.Fatalf("unexpected list item: %+v", resp.List[0])
	}
}

func TestUpdateChannelIntegrationTemplateStatusRejectsArchivedReenable(t *testing.T) {
	svcCtx := newChannelIntegrationTemplateTestSvc(t)
	if err := svcCtx.DB.Exec(`INSERT INTO sys_channel_integration_template (
		template_code, template_name, template_type, target_code, scope_type, platform_id, tenant_id, merchant_id, status,
		metadata_config, secret_ref_config, intent_contract_config, impact_scope_config, remark, create_by, create_time, update_by
	) VALUES (
		'tpl_archived', '归档模板', 'integration', 'sms_provider', 'platform', 1, 0, 0, 'archived',
		'{}', '{"accessKeySecretRef":"credential://sys/integration/sms/access-key-secret"}', '{}', '{}', '', 'seed', CURRENT_TIMESTAMP, 'seed'
	)`).Error; err != nil {
		t.Fatalf("seed archived template failed: %v", err)
	}

	logic := NewUpdateChannelIntegrationTemplateStatusLogic(context.Background(), svcCtx)
	_, err := logic.UpdateChannelIntegrationTemplateStatus(&sysclient.UpdateChannelIntegrationTemplateStatusReq{
		Ids:      []int64{1},
		Status:   channeltemplate.StatusEnabled,
		UpdateBy: "tester",
	})
	if err == nil {
		t.Fatal("expected archived -> enabled transition to fail")
	}
	if !strings.Contains(err.Error(), "归档") {
		t.Fatalf("unexpected transition error: %v", err)
	}
}
