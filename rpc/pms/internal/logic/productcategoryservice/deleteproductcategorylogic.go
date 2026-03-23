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
	if err := l.svcCtx.DB.WithContext(l.ctx).Table("pms_product_category").Where("id IN ?", in.Ids).Updates(map[string]interface{}{"is_deleted": 1, "update_by": in.UpdateBy, "update_time": time.Now()}).Error; err != nil {
		logc.Errorf(l.ctx, "删除产品分类失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("删除产品分类失败")
	}
	return &pmsclient.DeleteProductCategoryResp{}, nil
}
