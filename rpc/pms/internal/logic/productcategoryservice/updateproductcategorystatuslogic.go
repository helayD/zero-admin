package productcategoryservicelogic

import (
	"context"
	"errors"
	"strconv"
	"time"

	logiccommon "github.com/feihua/zero-admin/rpc/pms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateProductCategoryStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateProductCategoryStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProductCategoryStatusLogic {
	return &UpdateProductCategoryStatusLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}
func (l *UpdateProductCategoryStatusLogic) UpdateProductCategoryStatus(in *pmsclient.UpdateProductCategoryStatusReq) (*pmsclient.UpdateProductCategoryStatusResp, error) {
	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, in.UpdateBy)
	if err != nil {
		return nil, err
	}
	if _, err := logiccommon.EnsureCategoryScope(l.ctx, l.svcCtx.DB, currentScope, in.Ids, "pms.product_category.status", in.UpdateBy, "", "status="+strconv.Itoa(int(in.Status))); err != nil {
		return nil, err
	}
	if err := l.svcCtx.DB.WithContext(l.ctx).Table("pms_product_category").Where("id IN ?", in.Ids).Updates(map[string]interface{}{"is_enabled": in.Status, "update_by": in.UpdateBy, "update_time": time.Now()}).Error; err != nil {
		logc.Errorf(l.ctx, "更新产品分类状态失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("更新产品分类状态失败")
	}
	return &pmsclient.UpdateProductCategoryStatusResp{}, nil
}
