package productbrandservicelogic

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

type DeleteProductBrandLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteProductBrandLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteProductBrandLogic {
	return &DeleteProductBrandLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}
func (l *DeleteProductBrandLogic) DeleteProductBrand(in *pmsclient.DeleteProductBrandReq) (*pmsclient.DeleteProductBrandResp, error) {
	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, in.UpdateBy)
	if err != nil {
		return nil, err
	}
	if _, err := logiccommon.EnsureBrandScope(l.ctx, l.svcCtx.DB, currentScope, in.Ids, "pms.product_brand.delete", in.UpdateBy, "", "delete brand"); err != nil {
		return nil, err
	}
	if err := l.svcCtx.DB.WithContext(l.ctx).Table("pms_product_brand").Where("id IN ?", in.Ids).Updates(map[string]interface{}{"is_deleted": 1, "update_by": in.UpdateBy, "update_time": time.Now()}).Error; err != nil {
		logc.Errorf(l.ctx, "删除商品品牌失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("删除商品品牌失败")
	}
	return &pmsclient.DeleteProductBrandResp{}, nil
}
