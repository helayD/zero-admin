package productcategoryservicelogic

import (
	"context"
	"errors"
	"fmt"
	"strings"

	logiccommon "github.com/feihua/zero-admin/rpc/pms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type AddProductCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

type productCategoryCreateRow struct {
	ID           int64  `gorm:"column:id"`
	ParentID     int64  `gorm:"column:parent_id"`
	Name         string `gorm:"column:name"`
	Level        int32  `gorm:"column:level"`
	ProductCount int32  `gorm:"column:product_count"`
	ProductUnit  string `gorm:"column:product_unit"`
	NavStatus    int32  `gorm:"column:nav_status"`
	Sort         int32  `gorm:"column:sort"`
	Icon         string `gorm:"column:icon"`
	Keywords     string `gorm:"column:keywords"`
	Description  string `gorm:"column:description"`
	IsEnabled    int32  `gorm:"column:is_enabled"`
	CreateBy     int64  `gorm:"column:create_by"`
	PlatformID   int64  `gorm:"column:platform_id"`
	TenantID     int64  `gorm:"column:tenant_id"`
	MerchantID   int64  `gorm:"column:merchant_id"`
	IsDeleted    int32  `gorm:"column:is_deleted"`
}

func NewAddProductCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddProductCategoryLogic {
	return &AddProductCategoryLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *AddProductCategoryLogic) AddProductCategory(in *pmsclient.AddProductCategoryReq) (*pmsclient.AddProductCategoryResp, error) {
	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, in.CreateBy)
	if err != nil {
		return nil, err
	}
	var count int64
	if err := l.svcCtx.DB.WithContext(l.ctx).Table("pms_product_category").Where("is_deleted = 0 AND parent_id = ? AND name = ? AND platform_id = ? AND tenant_id = ? AND merchant_id = ?", in.ParentId, strings.TrimSpace(in.Name), currentScope.PlatformID, currentScope.TenantID, currentScope.MerchantID).Count(&count).Error; err != nil {
		logc.Errorf(l.ctx, "校验商品分类重复失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("校验商品分类重复失败")
	}
	if count > 0 {
		return nil, errors.New(fmt.Sprintf("商品分类名称：%s,已存在", in.Name))
	}

	item := &productCategoryCreateRow{
		ParentID:     in.ParentId,
		Name:         strings.TrimSpace(in.Name),
		Level:        in.Level,
		ProductCount: 0,
		ProductUnit:  in.ProductUnit,
		NavStatus:    in.NavStatus,
		Sort:         in.Sort,
		Icon:         in.Icon,
		Keywords:     in.Keywords,
		Description:  in.Description,
		IsEnabled:    in.IsEnabled,
		CreateBy:     in.CreateBy,
		PlatformID:   currentScope.PlatformID,
		TenantID:     currentScope.TenantID,
		MerchantID:   currentScope.MerchantID,
		IsDeleted:    0,
	}

	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("pms_product_category").Select("*").Create(item).Error; err != nil {
			return err
		}
		return replaceCategoryRelations(tx, item.ID, in.ProductAttributeIdList, currentScope)
	})
	if err != nil {
		if isCategoryBindingValidationError(err) {
			return nil, err
		}
		logc.Errorf(l.ctx, "添加产品分类失败,参数:%+v,异常:%s", item, err.Error())
		return nil, errors.New("添加产品分类失败")
	}
	return &pmsclient.AddProductCategoryResp{Pong: "ok"}, nil
}
