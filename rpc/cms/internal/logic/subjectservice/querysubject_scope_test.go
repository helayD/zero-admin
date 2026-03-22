package subjectservicelogic

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/cms/cmsclient"
	"github.com/feihua/zero-admin/rpc/cms/gen/model"
	"github.com/feihua/zero-admin/rpc/cms/internal/svc"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newSubjectScopeTestSvc(t *testing.T) *svc.ServiceContext {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "subject-scope.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&model.CmsSubject{}); err != nil {
		t.Fatalf("auto migrate subject failed: %v", err)
	}
	for _, stmt := range []string{
		`ALTER TABLE cms_subject ADD COLUMN platform_id INTEGER NOT NULL DEFAULT 1`,
		`ALTER TABLE cms_subject ADD COLUMN tenant_id INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE cms_subject ADD COLUMN merchant_id INTEGER NOT NULL DEFAULT 0`,
	} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("alter subject scope columns failed: %v", err)
		}
	}

	return &svc.ServiceContext{DB: db}
}

func seedSubject(t *testing.T, db *gorm.DB, item model.CmsSubject, current pkgscope.GovernanceScope) {
	t.Helper()

	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("seed subject failed: %v", err)
	}
	if err := db.Exec(
		`UPDATE cms_subject SET platform_id = ?, tenant_id = ?, merchant_id = ? WHERE id = ?`,
		current.PlatformID,
		current.TenantID,
		current.MerchantID,
		item.ID,
	).Error; err != nil {
		t.Fatalf("update subject scope failed: %v", err)
	}
}

func TestQuerySubjectListFiltersByGovernanceScope(t *testing.T) {
	svcCtx := newSubjectScopeTestSvc(t)
	now := time.Now()

	merchantScope, err := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 301)
	if err != nil {
		t.Fatalf("normalize merchant scope failed: %v", err)
	}
	otherScope, err := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 302)
	if err != nil {
		t.Fatalf("normalize other scope failed: %v", err)
	}

	seedSubject(t, svcCtx.DB, model.CmsSubject{
		ID:              1,
		CategoryID:      100,
		Title:           "商户 301 专题",
		Pic:             "301.png",
		ProductCount:    3,
		RecommendStatus: 1,
		CollectCount:    1,
		ReadCount:       10,
		CommentCount:    0,
		AlbumPics:       "301.png",
		Description:     "scope 301",
		ShowStatus:      1,
		Content:         "scope 301",
		ForwardCount:    0,
		CategoryName:    "运营专题",
		Sort:            1,
		CreateBy:        "seed",
		CreateTime:      now,
		UpdateBy:        "seed",
	}, merchantScope)
	seedSubject(t, svcCtx.DB, model.CmsSubject{
		ID:              2,
		CategoryID:      100,
		Title:           "商户 302 专题",
		Pic:             "302.png",
		ProductCount:    2,
		RecommendStatus: 1,
		CollectCount:    1,
		ReadCount:       5,
		CommentCount:    0,
		AlbumPics:       "302.png",
		Description:     "scope 302",
		ShowStatus:      1,
		Content:         "scope 302",
		ForwardCount:    0,
		CategoryName:    "运营专题",
		Sort:            1,
		CreateBy:        "seed",
		CreateTime:      now,
		UpdateBy:        "seed",
	}, otherScope)

	logic := NewQuerySubjectListLogic(context.Background(), svcCtx)
	resp, err := logic.QuerySubjectList(&cmsclient.QuerySubjectListReq{
		PageNum:         1,
		PageSize:        10,
		RecommendStatus: 2,
		ShowStatus:      2,
		Scope: &cmsclient.GovernanceScope{
			ScopeType:  merchantScope.ScopeType,
			PlatformId: merchantScope.PlatformID,
			TenantId:   merchantScope.TenantID,
			MerchantId: merchantScope.MerchantID,
		},
	})
	if err != nil {
		t.Fatalf("query subject list failed: %v", err)
	}
	if resp.Total != 1 || len(resp.List) != 1 {
		t.Fatalf("expected one scoped subject, got total=%d len=%d", resp.Total, len(resp.List))
	}
	if resp.List[0].Id != 1 {
		t.Fatalf("expected subject 1, got %+v", resp.List[0])
	}
}

func TestQuerySubjectDetailRejectsCrossScopeLookup(t *testing.T) {
	svcCtx := newSubjectScopeTestSvc(t)
	now := time.Now()

	tenantScope, err := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeTenant, 1, 10, 0)
	if err != nil {
		t.Fatalf("normalize tenant scope failed: %v", err)
	}
	otherScope, err := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeTenant, 1, 20, 0)
	if err != nil {
		t.Fatalf("normalize other scope failed: %v", err)
	}

	seedSubject(t, svcCtx.DB, model.CmsSubject{
		ID:              11,
		CategoryID:      101,
		Title:           "tenant 10 专题",
		Pic:             "10.png",
		ProductCount:    1,
		RecommendStatus: 1,
		CollectCount:    0,
		ReadCount:       1,
		CommentCount:    0,
		AlbumPics:       "10.png",
		Description:     "tenant 10",
		ShowStatus:      1,
		Content:         "tenant 10",
		ForwardCount:    0,
		CategoryName:    "租户专题",
		Sort:            1,
		CreateBy:        "seed",
		CreateTime:      now,
		UpdateBy:        "seed",
	}, tenantScope)

	logic := NewQuerySubjectDetailLogic(context.Background(), svcCtx)
	_, err = logic.QuerySubjectDetail(&cmsclient.QuerySubjectDetailReq{
		Id: 11,
		Scope: &cmsclient.GovernanceScope{
			ScopeType:  otherScope.ScopeType,
			PlatformId: otherScope.PlatformID,
			TenantId:   otherScope.TenantID,
			MerchantId: otherScope.MerchantID,
		},
	})
	if err == nil {
		t.Fatal("expected cross-scope subject detail to fail")
	}
	if !strings.Contains(err.Error(), "专题不存在") {
		t.Fatalf("expected not found error, got %v", err)
	}
}
