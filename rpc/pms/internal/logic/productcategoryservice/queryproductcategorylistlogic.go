package productcategoryservicelogic

import (
	"context"
	"errors"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/pkg/time_util"
	logiccommon "github.com/feihua/zero-admin/rpc/pms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

type QueryProductCategoryListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryProductCategoryListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryProductCategoryListLogic {
	return &QueryProductCategoryListLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}
func (l *QueryProductCategoryListLogic) QueryProductCategoryList(in *pmsclient.QueryProductCategoryListReq) (*pmsclient.QueryProductCategoryListResp, error) {
	current, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, err
	}
	db := l.svcCtx.DB.WithContext(l.ctx).Table("pms_product_category").Where("is_deleted = 0")
	if in.Name != "" {
		db = db.Where("name LIKE ?", "%"+in.Name+"%")
	}
	if in.NavStatus != 2 {
		db = db.Where("nav_status = ?", in.NavStatus)
	}
	if in.Keywords != "" {
		db = db.Where("keywords LIKE ?", "%"+in.Keywords+"%")
	}
	if in.IsEnabled != 2 {
		db = db.Where("is_enabled = ?", in.IsEnabled)
	}
	if in.ParentId != 1000 {
		db = db.Where("parent_id = ?", in.ParentId)
	}
	scopeWhere, scopeArgs := pkgscope.ScopeFilterSQL("", current)
	db = db.Where(scopeWhere, scopeArgs...)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		logc.Errorf(l.ctx, "查询产品分类总数失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("查询产品分类列表失败")
	}
	var rows []logiccommon.CatalogScopeRow
	if err := db.Order("sort asc, id desc").Offset(int((in.PageNum - 1) * in.PageSize)).Limit(int(in.PageSize)).Scan(&rows).Error; err != nil {
		logc.Errorf(l.ctx, "查询产品分类列表失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("查询产品分类列表失败")
	}
	list := make([]*pmsclient.ProductCategoryListData, 0, len(rows))
	for _, item := range rows {
		list = append(list, &pmsclient.ProductCategoryListData{Id: item.ID, ParentId: item.ParentID, Name: item.Name, Level: item.Level, ProductCount: item.ProductCount, ProductUnit: item.ProductUnit, NavStatus: item.NavStatus, Sort: item.Sort, Icon: item.Icon, Keywords: item.Keywords, Description: item.Description, IsEnabled: item.IsEnabled, CreateBy: item.CreateBy, CreateTime: time_util.TimeToStr(item.CreateTime), UpdateBy: derefCategoryInt64(item.UpdateBy), UpdateTime: time_util.TimeToString(item.UpdateTime), ScopeType: item.GovernanceScope().ScopeType, PlatformId: item.PlatformID, TenantId: item.TenantID, MerchantId: item.MerchantID})
	}
	return &pmsclient.QueryProductCategoryListResp{Total: total, List: list}, nil
}
func derefCategoryInt64(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}
