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
	"gorm.io/gorm"
)

type QueryProductCategoryDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryProductCategoryDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryProductCategoryDetailLogic {
	return &QueryProductCategoryDetailLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}
func (l *QueryProductCategoryDetailLogic) QueryProductCategoryDetail(in *pmsclient.QueryProductCategoryDetailReq) (*pmsclient.QueryProductCategoryDetailResp, error) {
	current, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, err
	}
	scopeWhere, scopeArgs := pkgscope.ScopeFilterSQL("", current)
	var item logiccommon.CatalogScopeRow
	err = l.svcCtx.DB.WithContext(l.ctx).Table("pms_product_category").Where("id = ? AND is_deleted = 0", in.Id).Where(scopeWhere, scopeArgs...).Take(&item).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		logc.Errorf(l.ctx, "产品分类不存在, 请求参数：%+v, 异常信息: %s", in, err.Error())
		return nil, errors.New("产品分类不存在")
	case err != nil:
		logc.Errorf(l.ctx, "查询产品分类异常, 请求参数：%+v, 异常信息: %s", in, err.Error())
		return nil, errors.New("查询产品分类异常")
	}
	return &pmsclient.QueryProductCategoryDetailResp{Id: item.ID, ParentId: item.ParentID, Name: item.Name, Level: item.Level, ProductCount: item.ProductCount, ProductUnit: item.ProductUnit, NavStatus: item.NavStatus, Sort: item.Sort, Icon: item.Icon, Keywords: item.Keywords, Description: item.Description, IsEnabled: item.IsEnabled, CreateBy: item.CreateBy, CreateTime: time_util.TimeToStr(item.CreateTime), UpdateBy: derefCategoryInt64(item.UpdateBy), UpdateTime: time_util.TimeToString(item.UpdateTime), ScopeType: item.GovernanceScope().ScopeType, PlatformId: item.PlatformID, TenantId: item.TenantID, MerchantId: item.MerchantID}, nil
}
