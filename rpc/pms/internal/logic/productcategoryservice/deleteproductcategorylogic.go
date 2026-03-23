package productcategoryservicelogic

import (
	"context"
	"errors"
	"time"

	logiccommon "github.com/feihua/zero-admin/rpc/pms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteProductCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteProductCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteProductCategoryLogic {
	return &DeleteProductCategoryLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *DeleteProductCategoryLogic) DeleteProductCategory(in *pmsclient.DeleteProductCategoryReq) (*pmsclient.DeleteProductCategoryResp, error) {
	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, in.UpdateBy)
	if err != nil {
		return nil, err
	}
	if _, err := logiccommon.EnsureCategoryScope(l.ctx, l.svcCtx.DB, currentScope, in.Ids, "pms.product_category.delete", in.UpdateBy, "", "delete category"); err != nil {
		return nil, err
	}
	if err := logiccommon.EnsureCatalogDeleteAllowed(l.ctx, l.svcCtx.DB, currentScope, in.Ids, []logiccommon.CatalogReferenceCheck{
		{Table: "pms_product_category", Column: "parent_id", Message: "商品分类下仍有子分类，无法删除", ScopeAware: true},
		{Table: "pms_product_attribute_group", Column: "category_id", Message: "商品分类已被属性分组引用，无法删除", ScopeAware: true},
		{Table: "pms_product_spec", Column: "category_id", Message: "商品分类已被商品规格引用，无法删除", ScopeAware: true},
		{Table: "pms_product_category_attribute_relation", Column: "product_category_id", Message: "商品分类已绑定商品属性关系，无法删除", ScopeAware: true},
		{Table: "pms_product_spu", Column: "category_id", Message: "商品分类已被商品建档引用，无法删除", ScopeAware: true},
	}); err != nil {
		return nil, err
	}
	if err := l.svcCtx.DB.WithContext(l.ctx).Table("pms_product_category").Where("id IN ?", in.Ids).Updates(map[string]interface{}{"is_deleted": 1, "update_by": in.UpdateBy, "update_time": time.Now()}).Error; err != nil {
		logc.Errorf(l.ctx, "删除产品分类失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("删除产品分类失败")
	}
	return &pmsclient.DeleteProductCategoryResp{}, nil
}
