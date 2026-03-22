package preferredareaproductrelationservicelogic

import (
	"context"
	"errors"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/cms/cmsclient"
	"github.com/feihua/zero-admin/rpc/cms/internal/svc"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

// QueryPreferredAreaProductRelationListLogic 查询优选专区和产品关系列表
/*
Author: LiuFeiHua
Date: 2024/6/11 16:40
*/
type QueryPreferredAreaProductRelationListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

type preferredAreaRelationScopeRow struct {
	PlatformID int64 `gorm:"column:platform_id"`
	TenantID   int64 `gorm:"column:tenant_id"`
	MerchantID int64 `gorm:"column:merchant_id"`
}

func NewQueryPreferredAreaProductRelationListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryPreferredAreaProductRelationListLogic {
	return &QueryPreferredAreaProductRelationListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// QueryPreferredAreaProductRelationList 查询优选专区和产品关系列表
func (l *QueryPreferredAreaProductRelationListLogic) QueryPreferredAreaProductRelationList(in *cmsclient.QueryPreferredAreaProductRelationListReq) (*cmsclient.QueryPreferredAreaProductRelationListResp, error) {
	var productRow preferredAreaRelationScopeRow
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Table("pms_product_spu").
		Select("platform_id, tenant_id, merchant_id").
		Where("id = ?", in.ProductId).
		Take(&productRow).Error; err != nil {
		logc.Errorf(l.ctx, "查询优选专区和产品关系列表失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("查询优选专区和产品关系列表失败")
	}

	current, err := pkgscope.NormalizeGovernanceScope("", productRow.PlatformID, productRow.TenantID, productRow.MerchantID)
	if err != nil {
		logc.Errorf(l.ctx, "解析商品scope失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("查询优选专区和产品关系列表失败")
	}

	var ids []int64
	err = l.svcCtx.DB.WithContext(l.ctx).
		Table("cms_preferred_area_product_relation").
		Select("preferred_area_id").
		Where("product_id = ? AND platform_id = ? AND tenant_id = ? AND merchant_id = ?", in.ProductId, current.PlatformID, current.TenantID, current.MerchantID).
		Scan(&ids).Error

	if err != nil {
		logc.Errorf(l.ctx, "查询优选专区和产品关系列表失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("查询优选专区和产品关系列表失败")
	}

	return &cmsclient.QueryPreferredAreaProductRelationListResp{
		PreferredAreaIds: ids,
	}, nil
}
