package svc

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/feihua/zero-admin/pkg/sms"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newSmsConfigResolverTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "sms-config-resolver-test.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.Exec(`CREATE TABLE sys_system_config (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		config_group TEXT NOT NULL,
		config_key TEXT NOT NULL,
		config_value TEXT,
		value_type TEXT NOT NULL DEFAULT 'string',
		is_secret INTEGER NOT NULL DEFAULT 0,
		remark TEXT NOT NULL DEFAULT '',
		create_by TEXT NOT NULL DEFAULT 'system',
		create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
		update_by TEXT NOT NULL DEFAULT '',
		update_time DATETIME NULL,
		is_deleted INTEGER NOT NULL DEFAULT 1,
		UNIQUE(config_group, config_key)
	)`).Error; err != nil {
		t.Fatalf("create sys_system_config failed: %v", err)
	}
	return db
}

func seedSMSConfig(t *testing.T, db *gorm.DB, key, value string) {
	t.Helper()
	if err := db.Exec(`INSERT INTO sys_system_config (
		config_group, config_key, config_value, value_type, is_secret, remark, create_by, update_by, is_deleted
	) VALUES ('sms', ?, ?, 'string', 0, '', 'test', 'test', 1)`, key, value).Error; err != nil {
		t.Fatalf("seed sms config %s failed: %v", key, err)
	}
}

func TestSmsConfigResolverPrefersSystemConfig(t *testing.T) {
	db := newSmsConfigResolverTestDB(t)
	seedSMSConfig(t, db, "enabled", "true")
	seedSMSConfig(t, db, "provider", "aliyun")
	seedSMSConfig(t, db, "endpoint", "dysmsapi.aliyuncs.com")
	seedSMSConfig(t, db, "accessKeyId", "ak-from-db")
	seedSMSConfig(t, db, "accessKeySecret", "sk-from-db")
	seedSMSConfig(t, db, "signName", "杭州山河集信息科技")
	seedSMSConfig(t, db, "templateCode", "SMS_335160513")

	resolver := NewSmsConfigResolver(db, nil)
	got, err := resolver.Resolve(context.Background(), sms.SceneMemberLogin)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}

	if got.ProviderCode != sms.AliyunProviderCode {
		t.Fatalf("ProviderCode = %s, want %s", got.ProviderCode, sms.AliyunProviderCode)
	}
	if got.AccessKeyID != "ak-from-db" || got.AccessKeySecret != "sk-from-db" {
		t.Fatalf("credential = %s/%s, want db credential", got.AccessKeyID, got.AccessKeySecret)
	}
	if got.SignName != "杭州山河集信息科技" || got.TemplateCode != "SMS_335160513" {
		t.Fatalf("unexpected sign/template: %+v", got)
	}
	if got.Endpoint != "dysmsapi.aliyuncs.com" {
		t.Fatalf("Endpoint = %s", got.Endpoint)
	}
	if got.TimeoutSeconds != 5 || got.ExpireSeconds != 300 {
		t.Fatalf("unexpected default timeout/expire: %+v", got)
	}
}

func TestSmsConfigResolverSystemConfigMissingSecretFailsFast(t *testing.T) {
	db := newSmsConfigResolverTestDB(t)
	seedSMSConfig(t, db, "enabled", "true")
	seedSMSConfig(t, db, "provider", "aliyun")
	seedSMSConfig(t, db, "signName", "杭州山河集信息科技")
	seedSMSConfig(t, db, "templateCode", "SMS_335160513")

	resolver := NewSmsConfigResolver(db, nil)
	_, err := resolver.Resolve(context.Background(), sms.SceneMemberLogin)
	if err == nil {
		t.Fatal("expected missing AccessKey to fail")
	}
	if !strings.Contains(err.Error(), "AccessKey") {
		t.Fatalf("error should mention AccessKey, got %v", err)
	}
}

func TestSmsConfigResolverFallsBackWhenSystemConfigDisabled(t *testing.T) {
	db := newSmsConfigResolverTestDB(t)
	seedSMSConfig(t, db, "enabled", "false")

	resolver := NewSmsConfigResolver(db, nil)
	_, err := resolver.Resolve(context.Background(), sms.SceneMemberLogin)
	if err == nil {
		t.Fatal("expected fallback to missing template client to fail")
	}
	if !strings.Contains(err.Error(), "未启用") {
		t.Fatalf("unexpected error: %v", err)
	}
}
