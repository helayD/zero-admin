package productcategoryservicelogic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	logiccommon "github.com/feihua/zero-admin/rpc/pms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type UpdateProductCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateProductCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProductCategoryLogic {
	return &UpdateProductCategoryLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}
func (l *UpdateProductCategoryLogic) UpdateProductCategory(in *pmsclient.UpdateProductCategoryReq) (*pmsclient.UpdateProductCategoryResp, error) {
	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, in.UpdateBy)
	if err != nil {
		return nil, err
	}
	if _, err := logiccommon.EnsureCategoryScope(l.ctx, l.svcCtx.DB, currentScope, []int64{in.Id}, "pms.product_category.update", in.UpdateBy, "", fmt.Sprintf("categoryId=%d", in.Id)); err != nil {
		return nil, err
	}
	var count int64
	if err := l.svcCtx.DB.WithContext(l.ctx).Table("pms_product_category").Where("is_deleted = 0 AND id <> ? AND parent_id = ? AND name = ? AND platform_id = ? AND tenant_id = ? AND merchant_id = ?", in.Id, in.ParentId, strings.TrimSpace(in.Name), currentScope.PlatformID, currentScope.TenantID, currentScope.MerchantID).Count(&count).Error; err != nil {
		logc.Errorf(l.ctx, "校验商品分类重复失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("校验商品分类重复失败")
	}
	if count > 0 {
		return nil, errors.New(fmt.Sprintf("商品分类名称：%s,已存在", in.Name))
	}
	updates := map[string]interface{}{"parent_id": in.ParentId, "name": strings.TrimSpace(in.Name), "level": in.Level, "product_unit": in.ProductUnit, "nav_status": in.NavStatus, "sort": in.Sort, "icon": in.Icon, "keywords": in.Keywords, "description": in.Description, "is_enabled": in.IsEnabled, "update_by": in.UpdateBy, "update_time": time.Now()}
	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("pms_product_category").Where("id = ?", in.Id).Updates(updates).Error; err != nil {
			return err
		}
		return replaceCategoryRelations(tx, in.Id, in.ProductAttributeIdList, currentScope)
	})
	if err != nil {
		if isCategoryBindingValidationError(err) {
			return nil, err
		}
		logc.Errorf(l.ctx, "更新产品分类失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("更新产品分类失败")
	}
	return &pmsclient.UpdateProductCategoryResp{}, nil
}
