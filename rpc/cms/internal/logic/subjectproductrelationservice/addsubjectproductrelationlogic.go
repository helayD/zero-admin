package subjectproductrelationservicelogic

import (
	"context"
	"errors"
	"fmt"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/cms/gen/query"
	"github.com/zeromicro/go-zero/core/logc"

	"github.com/feihua/zero-admin/rpc/cms/cmsclient"
	"github.com/feihua/zero-admin/rpc/cms/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// AddSubjectProductRelationLogic 添加专题商品关系
/*
Author: LiuFeiHua
Date: 2024/6/11 16:41
*/
type AddSubjectProductRelationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

type businessScopeRow struct {
	ID         int64 `gorm:"column:id"`
	PlatformID int64 `gorm:"column:platform_id"`
	TenantID   int64 `gorm:"column:tenant_id"`
	MerchantID int64 `gorm:"column:merchant_id"`
}

func NewAddSubjectProductRelationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddSubjectProductRelationLogic {
	return &AddSubjectProductRelationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AddSubjectProductRelation 添加专题商品关系
func (l *AddSubjectProductRelationLogic) AddSubjectProductRelation(in *cmsclient.AddSubjectProductRelationReq) (*cmsclient.AddSubjectProductRelationResp, error) {
	err := l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		var productRow businessScopeRow
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

		subjectIDs := make([]int64, 0, len(in.SubjectId))
		seen := make(map[int64]struct{}, len(in.SubjectId))
		for _, id := range in.SubjectId {
			if id <= 0 {
				continue
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			subjectIDs = append(subjectIDs, id)
		}

		relationRows := make([]map[string]interface{}, 0, len(subjectIDs))
		if len(subjectIDs) > 0 {
			var subjectRows []businessScopeRow
			if err := tx.Table("cms_subject").
				Select("id, platform_id, tenant_id, merchant_id").
				Where("id in ?", subjectIDs).
				Find(&subjectRows).Error; err != nil {
				return err
			}
			if len(subjectRows) != len(subjectIDs) {
				return fmt.Errorf("存在无效专题，无法建立关联")
			}

			for _, row := range subjectRows {
				subjectScope, err := pkgscope.NormalizeGovernanceScope("", row.PlatformID, row.TenantID, row.MerchantID)
				if err != nil {
					return err
				}
				if !subjectScope.SameScope(productScope) {
					return fmt.Errorf("专题与商品不属于同一主体范围")
				}

				relationRows = append(relationRows, map[string]interface{}{
					"platform_id": productScope.PlatformID,
					"tenant_id":   productScope.TenantID,
					"merchant_id": productScope.MerchantID,
					"subject_id":  row.ID,
					"product_id":  in.ProductId,
				})
			}
		}

		if _, err := qtx.CmsSubjectProductRelation.WithContext(l.ctx).Where(qtx.CmsSubjectProductRelation.ProductID.Eq(in.ProductId)).Delete(); err != nil {
			return err
		}
		if len(relationRows) == 0 {
			return nil
		}

		return tx.Table("cms_subject_product_relation").Create(&relationRows).Error
	})
	if err != nil {
		logc.Errorf(l.ctx, "添加专题商品关系失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("添加专题商品关系失败")
	}

	return &cmsclient.AddSubjectProductRelationResp{}, nil
}
