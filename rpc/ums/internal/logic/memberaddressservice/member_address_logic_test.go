package memberaddressservicelogic

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

func newMemberAddressTestSvc(t *testing.T) *svc.ServiceContext {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "member-address.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err = db.AutoMigrate(&model.UmsMemberAddress{}); err != nil {
		t.Fatalf("auto migrate member address failed: %v", err)
	}
	query.SetDefault(db)
	return &svc.ServiceContext{DB: db}
}

func seedMemberAddress(t *testing.T, db *gorm.DB, item model.UmsMemberAddress) {
	t.Helper()
	if item.CreateTime.IsZero() {
		item.CreateTime = time.Now()
	}
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("seed member address failed: %v", err)
	}
}

func TestQueryMemberAddressListReturnsEmptyList(t *testing.T) {
	svcCtx := newMemberAddressTestSvc(t)
	logic := NewQueryMemberAddressListLogic(context.Background(), svcCtx)

	resp, err := logic.QueryMemberAddressList(&umsclient.QueryMemberAddressListReq{
		MemberId: 8,
		PageNum:  1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("QueryMemberAddressList returned error: %v", err)
	}
	if resp.List == nil {
		t.Fatalf("expected empty list instead of nil")
	}
	if len(resp.List) != 0 || resp.Total != 0 {
		t.Fatalf("expected empty result, got total=%d len=%d", resp.Total, len(resp.List))
	}
}

func TestAddMemberAddressMakesFirstAddressDefault(t *testing.T) {
	svcCtx := newMemberAddressTestSvc(t)
	logic := NewAddMemberAddressLogic(context.Background(), svcCtx)

	_, err := logic.AddMemberAddress(&umsclient.AddMemberAddressReq{
		MemberId:      4,
		ReceiverName:  "张三",
		ReceiverPhone: "16698129676",
		Province:      "广东省",
		City:          "深圳市",
		District:      "南山区",
		DetailAddress: "科技园测试路1号",
		PostalCode:    "518000",
		Tag:           "公司",
		IsDefault:     0,
	})
	if err != nil {
		t.Fatalf("AddMemberAddress returned error: %v", err)
	}

	var row model.UmsMemberAddress
	if err = svcCtx.DB.Where("member_id = ?", 4).First(&row).Error; err != nil {
		t.Fatalf("query inserted address failed: %v", err)
	}
	if row.IsDefault != 1 {
		t.Fatalf("expected first address to be default, got %d", row.IsDefault)
	}
}

func TestUpdateMemberAddressStatusKeepsSingleDefault(t *testing.T) {
	svcCtx := newMemberAddressTestSvc(t)
	now := time.Now()
	seedMemberAddress(t, svcCtx.DB, model.UmsMemberAddress{
		ID:            1,
		MemberID:      4,
		ReceiverName:  "张三",
		ReceiverPhone: "16698129676",
		Province:      "广东省",
		City:          "深圳市",
		District:      "南山区",
		DetailAddress: "科技园A座",
		PostalCode:    "518000",
		Tag:           "公司",
		IsDefault:     1,
		CreateTime:    now.Add(-time.Hour),
		IsDeleted:     0,
	})
	seedMemberAddress(t, svcCtx.DB, model.UmsMemberAddress{
		ID:            2,
		MemberID:      4,
		ReceiverName:  "李四",
		ReceiverPhone: "16698129677",
		Province:      "广东省",
		City:          "深圳市",
		District:      "福田区",
		DetailAddress: "福中路100号",
		PostalCode:    "518001",
		Tag:           "家",
		IsDefault:     0,
		CreateTime:    now,
		IsDeleted:     0,
	})

	logic := NewUpdateMemberAddressStatusLogic(context.Background(), svcCtx)
	_, err := logic.UpdateMemberAddressStatus(&umsclient.UpdateMemberAddressStatusReq{
		Id:        2,
		MemberId:  4,
		IsDefault: 1,
	})
	if err != nil {
		t.Fatalf("UpdateMemberAddressStatus returned error: %v", err)
	}

	var rows []model.UmsMemberAddress
	if err = svcCtx.DB.Where("member_id = ?", 4).Order("id ASC").Find(&rows).Error; err != nil {
		t.Fatalf("query addresses failed: %v", err)
	}
	if rows[0].IsDefault != 0 || rows[1].IsDefault != 1 {
		t.Fatalf("expected only address 2 default, got %+v", rows)
	}
}

func TestDeleteMemberAddressPromotesRemainingAddress(t *testing.T) {
	svcCtx := newMemberAddressTestSvc(t)
	now := time.Now()
	seedMemberAddress(t, svcCtx.DB, model.UmsMemberAddress{
		ID:            1,
		MemberID:      4,
		ReceiverName:  "张三",
		ReceiverPhone: "16698129676",
		Province:      "广东省",
		City:          "深圳市",
		District:      "南山区",
		DetailAddress: "科技园A座",
		PostalCode:    "518000",
		Tag:           "公司",
		IsDefault:     1,
		CreateTime:    now.Add(-time.Hour),
		IsDeleted:     0,
	})
	seedMemberAddress(t, svcCtx.DB, model.UmsMemberAddress{
		ID:            2,
		MemberID:      4,
		ReceiverName:  "李四",
		ReceiverPhone: "16698129677",
		Province:      "广东省",
		City:          "深圳市",
		District:      "福田区",
		DetailAddress: "福中路100号",
		PostalCode:    "518001",
		Tag:           "家",
		IsDefault:     0,
		CreateTime:    now,
		IsDeleted:     0,
	})

	logic := NewDeleteMemberAddressLogic(context.Background(), svcCtx)
	_, err := logic.DeleteMemberAddress(&umsclient.DeleteMemberAddressReq{
		Ids:      []int64{1},
		MemberId: 4,
	})
	if err != nil {
		t.Fatalf("DeleteMemberAddress returned error: %v", err)
	}

	var deleted model.UmsMemberAddress
	if err = svcCtx.DB.Where("id = ?", 1).First(&deleted).Error; err != nil {
		t.Fatalf("query deleted address failed: %v", err)
	}
	if deleted.IsDeleted != 1 || deleted.IsDefault != 0 {
		t.Fatalf("expected deleted default address to be hidden, got %+v", deleted)
	}

	var remaining model.UmsMemberAddress
	if err = svcCtx.DB.Where("id = ?", 2).First(&remaining).Error; err != nil {
		t.Fatalf("query remaining address failed: %v", err)
	}
	if remaining.IsDefault != 1 {
		t.Fatalf("expected remaining address promoted to default, got %d", remaining.IsDefault)
	}
}
