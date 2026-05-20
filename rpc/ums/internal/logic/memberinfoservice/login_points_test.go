package memberinfoservicelogic

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"google.golang.org/grpc/metadata"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/feihua/zero-admin/rpc/ums/gen/query"
	"github.com/feihua/zero-admin/rpc/ums/internal/svc"
	"github.com/feihua/zero-admin/rpc/ums/umsclient"
)

func newDailyLoginPointsTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "daily-login-points.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	stmts := []string{
		`CREATE TABLE ums_member_info (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			member_id INTEGER NOT NULL,
			level_id INTEGER NOT NULL DEFAULT 1,
			nickname TEXT NOT NULL DEFAULT '',
			mobile TEXT NOT NULL DEFAULT '',
			source INTEGER NOT NULL DEFAULT 1,
			avatar TEXT NOT NULL DEFAULT '',
			signature TEXT NOT NULL DEFAULT '',
			gender INTEGER NOT NULL DEFAULT 0,
			birthday DATETIME NULL,
			growth_point INTEGER NOT NULL DEFAULT 0,
			points INTEGER NOT NULL DEFAULT 0,
			total_points INTEGER NOT NULL DEFAULT 0,
			spend_amount REAL NOT NULL DEFAULT 0,
			order_count INTEGER NOT NULL DEFAULT 0,
			coupon_count INTEGER NOT NULL DEFAULT 0,
			comment_count INTEGER NOT NULL DEFAULT 0,
			return_count INTEGER NOT NULL DEFAULT 0,
			lottery_times INTEGER NOT NULL DEFAULT 0,
			last_login DATETIME NULL,
			is_enabled INTEGER NOT NULL DEFAULT 1,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NULL,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE oms_order_main (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE sms_coupon_record (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			member_id INTEGER NOT NULL,
			status INTEGER NOT NULL DEFAULT 0,
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE ums_member_identity (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			member_id INTEGER NOT NULL,
			real_name_status TEXT NOT NULL DEFAULT '',
			real_name_masked TEXT NOT NULL DEFAULT '',
			identity_no_masked TEXT NOT NULL DEFAULT '',
			provider_code TEXT NOT NULL DEFAULT '',
			credential_ref TEXT NOT NULL DEFAULT '',
			verified_at DATETIME NULL,
			failure_reason TEXT NOT NULL DEFAULT '',
			audit_status TEXT NOT NULL DEFAULT '',
			is_deleted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE ums_member_lottery_grant_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			member_id INTEGER NOT NULL,
			grant_type TEXT NOT NULL,
			grant_key TEXT NOT NULL,
			grant_times INTEGER NOT NULL DEFAULT 0,
			source_ref TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			create_time DATETIME NOT NULL,
			update_time DATETIME NULL,
			is_deleted INTEGER NOT NULL DEFAULT 0,
			UNIQUE(member_id, grant_type, grant_key, is_deleted)
		)`,
		`CREATE TABLE ums_member_points_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			member_id INTEGER NOT NULL,
			change_type INTEGER NOT NULL,
			change_points INTEGER NOT NULL,
			source_type INTEGER NOT NULL,
			description TEXT NOT NULL,
			operate_man TEXT NOT NULL,
			operate_note TEXT NOT NULL DEFAULT '',
			create_time DATETIME NOT NULL
		)`,
		`INSERT INTO ums_member_info (member_id, points, total_points, is_deleted)
		 VALUES (1001, 20, 30, 0)`,
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("exec test ddl failed: %v", err)
		}
	}
	query.SetDefault(db)
	return db
}

func TestQueryMemberInfoDetailGrantsDailyLoginPoints(t *testing.T) {
	db := newDailyLoginPointsTestDB(t)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(memberInfoGrantDailyLoginPointsMetadata, "true"))
	logic := NewQueryMemberInfoDetailLogic(ctx, &svc.ServiceContext{DB: db})

	for i := 0; i < 2; i++ {
		resp, err := logic.QueryMemberInfoDetail(&umsclient.QueryMemberInfoDetailReq{MemberId: 1001})
		if err != nil {
			t.Fatalf("QueryMemberInfoDetail returned error: %v", err)
		}
		if resp.Points != 30 || resp.TotalPoints != 40 {
			t.Fatalf("expected points=30,totalPoints=40, got points=%d,totalPoints=%d", resp.Points, resp.TotalPoints)
		}
	}

	var pointsLogCount int64
	if err := db.Table("ums_member_points_log").
		Where("member_id = ? AND operate_note LIKE ?", 1001, "daily_login:%").
		Count(&pointsLogCount).Error; err != nil {
		t.Fatalf("count points log failed: %v", err)
	}
	if pointsLogCount != 1 {
		t.Fatalf("expected one points log, got %d", pointsLogCount)
	}
}

func TestQueryMemberInfoDetailWithoutGrantMetadataDoesNotGrantDailyLoginPoints(t *testing.T) {
	db := newDailyLoginPointsTestDB(t)
	logic := NewQueryMemberInfoDetailLogic(context.Background(), &svc.ServiceContext{DB: db})

	resp, err := logic.QueryMemberInfoDetail(&umsclient.QueryMemberInfoDetailReq{MemberId: 1001})
	if err != nil {
		t.Fatalf("QueryMemberInfoDetail returned error: %v", err)
	}
	if resp.Points != 20 || resp.TotalPoints != 30 {
		t.Fatalf("expected points=20,totalPoints=30, got points=%d,totalPoints=%d", resp.Points, resp.TotalPoints)
	}

	var pointsLogCount int64
	if err := db.Table("ums_member_points_log").
		Where("member_id = ? AND operate_note LIKE ?", 1001, "daily_login:%").
		Count(&pointsLogCount).Error; err != nil {
		t.Fatalf("count points log failed: %v", err)
	}
	if pointsLogCount != 0 {
		t.Fatalf("expected no points log, got %d", pointsLogCount)
	}
}

func TestGrantDailyLoginPointsOnlyOncePerDay(t *testing.T) {
	db := newDailyLoginPointsTestDB(t)
	ctx := context.Background()
	now := time.Date(2026, 5, 20, 9, 30, 0, 0, time.Local)

	if err := grantDailyLoginPoints(ctx, db, 1001, now); err != nil {
		t.Fatalf("first grantDailyLoginPoints returned error: %v", err)
	}
	if err := grantDailyLoginPoints(ctx, db, 1001, now.Add(2*time.Hour)); err != nil {
		t.Fatalf("second grantDailyLoginPoints returned error: %v", err)
	}

	var member struct {
		Points      int32
		TotalPoints int32
	}
	if err := db.Table("ums_member_info").
		Select("points, total_points").
		Where("member_id = ?", 1001).
		Scan(&member).Error; err != nil {
		t.Fatalf("query member points failed: %v", err)
	}
	if member.Points != 30 || member.TotalPoints != 40 {
		t.Fatalf("expected points=30,total_points=40, got points=%d,total_points=%d", member.Points, member.TotalPoints)
	}

	var pointsLogCount int64
	if err := db.Table("ums_member_points_log").
		Where("member_id = ? AND operate_note = ?", 1001, "daily_login:2026-05-20").
		Count(&pointsLogCount).Error; err != nil {
		t.Fatalf("count points log failed: %v", err)
	}
	if pointsLogCount != 1 {
		t.Fatalf("expected one points log, got %d", pointsLogCount)
	}
}
