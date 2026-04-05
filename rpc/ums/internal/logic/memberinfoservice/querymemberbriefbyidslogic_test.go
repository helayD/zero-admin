package memberinfoservicelogic

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/feihua/zero-admin/rpc/ums/gen/model"
	"github.com/feihua/zero-admin/rpc/ums/gen/query"
	"github.com/feihua/zero-admin/rpc/ums/internal/svc"
	"github.com/feihua/zero-admin/rpc/ums/umsclient"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newMemberBriefTestSvc(t *testing.T) *svc.ServiceContext {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "member-brief.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&model.UmsMemberInfo{}); err != nil {
		t.Fatalf("auto migrate member info failed: %v", err)
	}
	query.SetDefault(db)
	return &svc.ServiceContext{DB: db}
}

func seedMemberBriefRow(t *testing.T, db *gorm.DB, item model.UmsMemberInfo) {
	t.Helper()
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("seed member info failed: %v", err)
	}
}

func TestQueryMemberBriefByIdsMasksSensitiveFields(t *testing.T) {
	svcCtx := newMemberBriefTestSvc(t)
	now := time.Now()
	seedMemberBriefRow(t, svcCtx.DB, model.UmsMemberInfo{
		ID:         1,
		MemberID:   1001,
		WxOpenid:   "wx-1001",
		LevelID:    9,
		Nickname:   "九克城用户",
		Mobile:     "13812345678",
		Source:     1,
		Password:   "pwd",
		Avatar:     "avatar.png",
		Signature:  "hello",
		Gender:     1,
		IsEnabled:  1,
		CreateTime: now,
		IsDeleted:  0,
	})

	logic := NewQueryMemberBriefByIdsLogic(context.Background(), svcCtx)
	resp, err := logic.QueryMemberBriefByIds(&umsclient.QueryMemberBriefByIdsReq{MemberIds: []int64{1001}})
	if err != nil {
		t.Fatalf("QueryMemberBriefByIds returned error: %v", err)
	}
	if len(resp.List) != 1 {
		t.Fatalf("expected one brief row, got %d", len(resp.List))
	}
	if resp.List[0].NicknameMasked != "九**户" {
		t.Fatalf("expected masked nickname, got %s", resp.List[0].NicknameMasked)
	}
	if resp.List[0].MobileMasked != "138****5678" {
		t.Fatalf("expected masked mobile, got %s", resp.List[0].MobileMasked)
	}
}

func TestQueryMemberBriefByIdsSupportsMissingRowsAndDedup(t *testing.T) {
	svcCtx := newMemberBriefTestSvc(t)
	now := time.Now()
	seedMemberBriefRow(t, svcCtx.DB, model.UmsMemberInfo{
		ID:         1,
		MemberID:   2002,
		WxOpenid:   "wx-2002",
		LevelID:    3,
		Nickname:   "张三",
		Mobile:     "13900001111",
		Source:     1,
		Password:   "pwd",
		Avatar:     "avatar.png",
		Signature:  "hello",
		Gender:     1,
		IsEnabled:  1,
		CreateTime: now,
		IsDeleted:  0,
	})
	seedMemberBriefRow(t, svcCtx.DB, model.UmsMemberInfo{
		ID:         2,
		MemberID:   3003,
		WxOpenid:   "wx-3003",
		LevelID:    5,
		Nickname:   "李四",
		Mobile:     "13722223333",
		Source:     2,
		Password:   "pwd",
		Avatar:     "avatar.png",
		Signature:  "hello",
		Gender:     2,
		IsEnabled:  1,
		CreateTime: now,
		IsDeleted:  0,
	})

	logic := NewQueryMemberBriefByIdsLogic(context.Background(), svcCtx)
	resp, err := logic.QueryMemberBriefByIds(&umsclient.QueryMemberBriefByIdsReq{MemberIds: []int64{3003, 2002, 3003, 9999}})
	if err != nil {
		t.Fatalf("QueryMemberBriefByIds returned error: %v", err)
	}
	if len(resp.List) != 2 {
		t.Fatalf("expected two matched brief rows, got %d", len(resp.List))
	}
	briefByID := make(map[int64]*umsclient.MemberBriefData, len(resp.List))
	for _, item := range resp.List {
		briefByID[item.MemberId] = item
	}
	if _, ok := briefByID[9999]; ok {
		t.Fatalf("missing member should not appear in response")
	}
	if briefByID[2002] == nil || briefByID[2002].NicknameMasked != "张*" {
		t.Fatalf("expected member 2002 to be masked and present, got %+v", briefByID[2002])
	}
	if briefByID[3003] == nil || briefByID[3003].MobileMasked != "137****3333" {
		t.Fatalf("expected member 3003 masked mobile, got %+v", briefByID[3003])
	}
}
