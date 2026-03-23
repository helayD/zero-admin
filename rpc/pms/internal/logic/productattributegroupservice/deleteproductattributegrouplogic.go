package productattributegroupservicelogic

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

type DeleteProductAttributeGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteProductAttributeGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteProductAttributeGroupLogic {
	return &DeleteProductAttributeGroupLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *DeleteProductAttributeGroupLogic) DeleteProductAttributeGroup(in *pmsclient.DeleteProductAttributeGroupReq) (*pmsclient.DeleteProductAttributeGroupResp, error) {
	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, in.UpdateBy)
	if err != nil {
		return nil, err
	}
	if _, err := logiccommon.EnsureAttributeGroupScope(l.ctx, l.svcCtx.DB, currentScope, in.Ids, "pms.product_attribute_group.delete", in.UpdateBy, "", "delete attribute group"); err != nil {
		return nil, err
	}
	if err := logiccommon.EnsureCatalogDeleteAllowed(l.ctx, l.svcCtx.DB, currentScope, in.Ids, []logiccommon.CatalogReferenceCheck{{
		Table:      "pms_product_attribute",
		Column:     "group_id",
		Message:    "商品属性分组已被商品属性引用，无法删除",
		ScopeAware: true,
	}}); err != nil {
		return nil, err
	}
	if err := l.svcCtx.DB.WithContext(l.ctx).Table("pms_product_attribute_group").Where("id IN ?", in.Ids).Updates(map[string]interface{}{"is_deleted": 1, "update_by": in.UpdateBy, "update_time": time.Now()}).Error; err != nil {
		logc.Errorf(l.ctx, "删除商品属性分组失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("删除商品属性分组失败")
	}
	return &pmsclient.DeleteProductAttributeGroupResp{Pong: "ok"}, nil
}
