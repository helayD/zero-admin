package productbrandservicelogic

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
)

// AddProductBrandLogic 添加商品品牌
type AddProductBrandLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddProductBrandLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddProductBrandLogic {
	return &AddProductBrandLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *AddProductBrandLogic) AddProductBrand(in *pmsclient.AddProductBrandReq) (*pmsclient.AddProductBrandResp, error) {
	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, in.CreateBy)
	if err != nil {
		return nil, err
	}

	var count int64
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Table("pms_product_brand").
		Where("is_deleted = 0 AND name = ? AND platform_id = ? AND tenant_id = ? AND merchant_id = ?", strings.TrimSpace(in.Name), currentScope.PlatformID, currentScope.TenantID, currentScope.MerchantID).
		Count(&count).Error; err != nil {
		logc.Errorf(l.ctx, "校验商品品牌重复失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("校验商品品牌重复失败")
	}
	if count > 0 {
		return nil, errors.New(fmt.Sprintf("品牌名称：%s,已存在", in.Name))
	}

	item := &logiccommon.CatalogScopeRow{
		Name:                strings.TrimSpace(in.Name),
		Logo:                in.Logo,
		BigPic:              in.BigPic,
		Description:         in.Description,
		FirstLetter:         strings.TrimSpace(in.FirstLetter),
		Sort:                in.Sort,
		RecommendStatus:     in.RecommendStatus,
		ProductCount:        in.ProductCount,
		ProductCommentCount: in.ProductCommentCount,
		IsEnabled:           in.IsEnabled,
		CreateBy:            in.CreateBy,
		PlatformID:          currentScope.PlatformID,
		TenantID:            currentScope.TenantID,
		MerchantID:          currentScope.MerchantID,
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Table("pms_product_brand").Create(item).Error; err != nil {
		logc.Errorf(l.ctx, "添加商品品牌失败,参数:%+v,异常:%s", item, err.Error())
		return nil, errors.New("添加商品品牌失败")
	}

	return &pmsclient.AddProductBrandResp{BrandId: item.ID}, nil
}
