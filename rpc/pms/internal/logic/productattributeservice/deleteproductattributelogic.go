package productattributeservicelogic

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

type DeleteProductAttributeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteProductAttributeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteProductAttributeLogic {
	return &DeleteProductAttributeLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *DeleteProductAttributeLogic) DeleteProductAttribute(in *pmsclient.DeleteProductAttributeReq) (*pmsclient.DeleteProductAttributeResp, error) {
	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, in.UpdateBy)
	if err != nil {
		return nil, err
	}
	if _, err := logiccommon.EnsureAttributeScope(l.ctx, l.svcCtx.DB, currentScope, in.Ids, "pms.product_attribute.delete", in.UpdateBy, "", "delete attribute"); err != nil {
		return nil, err
	}
	if err := logiccommon.EnsureCatalogDeleteAllowed(l.ctx, l.svcCtx.DB, currentScope, in.Ids, []logiccommon.CatalogReferenceCheck{
		{Table: "pms_product_category_attribute_relation", Column: "product_attribute_id", Message: "商品属性已被商品分类绑定，无法删除", ScopeAware: true},
		{Table: "pms_product_attribute_value", Column: "attribute_id", Message: "商品属性已被商品建档引用，无法删除", ScopeAware: false},
	}); err != nil {
		return nil, err
	}
	if err := l.svcCtx.DB.WithContext(l.ctx).Table("pms_product_attribute").Where("id IN ?", in.Ids).Updates(map[string]interface{}{"is_deleted": 1, "update_by": in.UpdateBy, "update_time": time.Now()}).Error; err != nil {
		logc.Errorf(l.ctx, "删除商品属性失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("删除商品属性失败")
	}
	return &pmsclient.DeleteProductAttributeResp{Pong: "ok"}, nil
}
