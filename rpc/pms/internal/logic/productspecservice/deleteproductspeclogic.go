package productspecservicelogic

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

type DeleteProductSpecLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteProductSpecLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteProductSpecLogic {
	return &DeleteProductSpecLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *DeleteProductSpecLogic) DeleteProductSpec(in *pmsclient.DeleteProductSpecReq) (*pmsclient.DeleteProductSpecResp, error) {
	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, in.UpdateBy)
	if err != nil {
		return nil, err
	}
	if _, err := logiccommon.EnsureSpecScope(l.ctx, l.svcCtx.DB, currentScope, in.Ids, "pms.product_spec.delete", in.UpdateBy, "", "delete spec"); err != nil {
		return nil, err
	}
	if err := logiccommon.EnsureCatalogDeleteAllowed(l.ctx, l.svcCtx.DB, currentScope, in.Ids, []logiccommon.CatalogReferenceCheck{{
		Table:      "pms_product_spec_value",
		Column:     "spec_id",
		Message:    "商品规格已被规格值引用，无法删除",
		ScopeAware: true,
	}}); err != nil {
		return nil, err
	}
	if err := l.svcCtx.DB.WithContext(l.ctx).Table("pms_product_spec").Where("id IN ?", in.Ids).Updates(map[string]interface{}{"is_deleted": 1, "update_by": in.UpdateBy, "update_time": time.Now()}).Error; err != nil {
		logc.Errorf(l.ctx, "删除商品规格失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("删除商品规格失败")
	}
	return &pmsclient.DeleteProductSpecResp{Pong: "ok"}, nil
}
