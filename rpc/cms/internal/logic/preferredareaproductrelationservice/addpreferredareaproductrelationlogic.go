package preferredareaproductrelationservicelogic

import (
	"context"
	"errors"
	"fmt"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/cms/cmsclient"
	"github.com/feihua/zero-admin/rpc/cms/gen/query"
	"github.com/feihua/zero-admin/rpc/cms/internal/svc"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// AddPreferredAreaProductRelationLogic 添加优选专区和产品关系
/*
Author: LiuFeiHua
Date: 2024/6/11 16:40
*/
type AddPreferredAreaProductRelationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

type preferredAreaScopeRow struct {
	ID         int64 `gorm:"column:id"`
	PlatformID int64 `gorm:"column:platform_id"`
	TenantID   int64 `gorm:"column:tenant_id"`
	MerchantID int64 `gorm:"column:merchant_id"`
}

func NewAddPreferredAreaProductRelationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddPreferredAreaProductRelationLogic {
	return &AddPreferredAreaProductRelationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AddPreferredAreaProductRelation 添加优选专区和产品关系
func (l *AddPreferredAreaProductRelationLogic) AddPreferredAreaProductRelation(in *cmsclient.AddPreferredAreaProductRelationReq) (*cmsclient.AddPreferredAreaProductRelationResp, error) {
	err := l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)

		var productRow preferredAreaScopeRow
		if err := tx.Table("pms_product_spu").
			Select("id, platform_id, tenant_id, merchant_id").
			Where("id = ?", in.ProductId).
			Take(&productRow).Error; err != nil {
			return err
		}

		productScope, err := pkgscope.NormalizeGovernanceScope("", productRow.PlatformID, productRow.TenantID, productRow.MerchantID)
		if err != nil {
			return err
		}

		preferredAreaIDs := make([]int64, 0, len(in.PreferredAreaId))
		seen := make(map[int64]struct{}, len(in.PreferredAreaId))
		for _, id := range in.PreferredAreaId {
			if id <= 0 {
				continue
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			preferredAreaIDs = append(preferredAreaIDs, id)
		}

		relationRows := make([]map[string]interface{}, 0, len(preferredAreaIDs))
		if len(preferredAreaIDs) > 0 {
			var preferredAreaRows []preferredAreaScopeRow
			if err := tx.Table("cms_preferred_area").
				Select("id, platform_id, tenant_id, merchant_id").
				Where("id in ?", preferredAreaIDs).
				Find(&preferredAreaRows).Error; err != nil {
				return err
			}
			if len(preferredAreaRows) != len(preferredAreaIDs) {
				return fmt.Errorf("存在无效优选专区，无法建立关联")
			}

			for _, row := range preferredAreaRows {
				preferredAreaScope, err := pkgscope.NormalizeGovernanceScope("", row.PlatformID, row.TenantID, row.MerchantID)
				if err != nil {
					return err
				}
				if !preferredAreaScope.SameScope(productScope) {
					return fmt.Errorf("优选专区与商品不属于同一主体范围")
				}

				relationRows = append(relationRows, map[string]interface{}{
					"platform_id":       productScope.PlatformID,
					"tenant_id":         productScope.TenantID,
					"merchant_id":       productScope.MerchantID,
					"preferred_area_id": row.ID,
					"product_id":        in.ProductId,
				})
			}
		}

		if _, err := qtx.CmsPreferredAreaProductRelation.WithContext(l.ctx).Where(qtx.CmsPreferredAreaProductRelation.ProductID.Eq(in.ProductId)).Delete(); err != nil {
			return err
		}
		if len(relationRows) == 0 {
			return nil
		}

		return tx.Table("cms_preferred_area_product_relation").Create(&relationRows).Error
	})
	if err != nil {
		logc.Errorf(l.ctx, "添加优选专区和产品关系失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("添加优选专区和产品关系失败")
	}

	return &cmsclient.AddPreferredAreaProductRelationResp{}, nil
}
