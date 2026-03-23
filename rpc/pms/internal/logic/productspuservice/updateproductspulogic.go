package productspuservicelogic

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/feihua/zero-admin/rpc/pms/gen/model"
	"github.com/feihua/zero-admin/rpc/pms/gen/query"
	logiccommon "github.com/feihua/zero-admin/rpc/pms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// UpdateProductSpuLogic 更新商品SPU
/*
Author: LiuFeiHua
Date: 2025/06/16 14:37:38
*/
type UpdateProductSpuLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateProductSpuLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProductSpuLogic {
	return &UpdateProductSpuLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateProductSpu 更新商品SPU
// 逻辑步骤：
// 1.更新商品基本信息
// 2.会员价格
// 3.阶梯价格
// 4.满减价格
// 5.更新sku库存信息
// 6.更新商品参数,添加自定义商品规格
func (l *UpdateProductSpuLogic) UpdateProductSpu(in *pmsclient.ProductSpuReq) (*pmsclient.ProductSpuResp, error) {
	spu := query.PmsProductSpu
	q := spu.WithContext(l.ctx)
	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, in.CreateBy)
	if err != nil {
		return nil, err
	}
	if _, err := logiccommon.EnsureProductScope(l.ctx, l.svcCtx.DB, currentScope, []int64{in.Id}, "pms.product_spu.update", in.CreateBy, "", "spuId="+strconv.FormatInt(in.Id, 10)); err != nil {
		return nil, err
	}

	// 1.根据商品SPUid查询商品SPU是否已存在
	detail, err := q.Where(spu.ID.Eq(in.Id)).First()

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		logc.Errorf(l.ctx, "商品SPU不存在, 请求参数：%+v, 异常信息: %s", in, err.Error())
		return nil, errors.New("商品SPU不存在")
	case err != nil:
		logc.Errorf(l.ctx, "查询商品SPU异常, 请求参数：%+v, 异常信息: %s", in, err.Error())
		return nil, errors.New("查询商品SPU异常")
	}

	now := time.Now()
	summary, err := logiccommon.ValidateProductDraft(l.ctx, l.svcCtx.DB, currentScope, in)
	if err != nil {
		return nil, err
	}
	item := &model.PmsProductSpu{
		ID:                  in.Id,
		ProductSn:           in.ProductSn,
		Name:                in.Name,
		CategoryID:          in.CategoryId,
		CategoryIds:         in.CategoryIds,
		CategoryName:        in.CategoryName,
		BrandID:             in.BrandId,
		BrandName:           in.BrandName,
		Unit:                in.Unit,
		Weight:              float64(in.Weight),
		Keywords:            in.Keywords,
		AlbumPics:           in.AlbumPics,
		MainPic:             in.MainPic,
		PriceRange:          summary.PriceRange,
		PublishStatus:       in.PublishStatus,
		NewStatus:           in.NewStatus,
		RecommendStatus:     in.RecommendStatus,
		VerifyStatus:        in.VerifyStatus,
		PreviewStatus:       in.PreviewStatus,
		Sort:                in.Sort,
		NewStatusSort:       in.NewStatusSort,
		RecommendStatusSort: in.RecommendStatusSort,
		Sales:               in.Sales,
		Stock:               summary.TotalStock,
		LowStock:            summary.LowStock,
		PromotionType:       in.PromotionType,
		SubTitle:            in.SubTitle,
		DetailHTML:          in.DetailHtml,
		DetailMobileHTML:    in.DetailMobileHtml,
		CreateBy:            detail.CreateBy,
		CreateTime:          detail.CreateTime,
		UpdateBy:            &in.CreateBy,
		UpdateTime:          &now,
	}

	spuId := in.Id

	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		if err := tx.Model(&model.PmsProductSpu{}).Where("id = ?", in.Id).Save(item).Error; err != nil {
			return err
		}

		memberPrice := qtx.PmsMemberPrice.WithContext(l.ctx)
		if _, err := memberPrice.Where(qtx.PmsMemberPrice.ProductID.Eq(spuId)).Delete(); err != nil {
			return err
		}
		for _, list := range in.MemberPriceList {
			if err := memberPrice.Create(&model.PmsMemberPrice{
				ProductID:       spuId,
				MemberLevelID:   list.LevelId,
				MemberPrice:     list.Price,
				MemberLevelName: list.LevelName,
			}); err != nil {
				return err
			}
		}

		ladder := qtx.PmsProductLadder.WithContext(l.ctx)
		if _, err := ladder.Where(qtx.PmsProductLadder.ProductID.Eq(spuId)).Delete(); err != nil {
			return err
		}
		for _, list := range in.ProductLadderList {
			if err := ladder.Create(&model.PmsProductLadder{
				ProductID: spuId,
				Count:     list.Count,
				Discount:  list.Discount,
				Price:     list.Price,
			}); err != nil {
				return err
			}
		}

		full := qtx.PmsProductFullReduction.WithContext(l.ctx)
		if _, err := full.Where(qtx.PmsProductFullReduction.ProductID.Eq(spuId)).Delete(); err != nil {
			return err
		}
		for _, list := range in.ProductFullReductionList {
			if err := full.Create(&model.PmsProductFullReduction{
				ProductID:   spuId,
				FullPrice:   list.FullPrice,
				ReducePrice: list.ReducePrice,
			}); err != nil {
				return err
			}
		}

		sku := qtx.PmsProductSku.WithContext(l.ctx)
		if _, err := sku.Where(qtx.PmsProductSku.SpuID.Eq(spuId)).Delete(); err != nil {
			return err
		}
		for _, list := range in.SkuStockList {
			skuCode, err := logiccommon.EnsureSkuCode(l.ctx, tx, currentScope, spuId, list.SkuCode, list.SpecData, list.Name, nil)
			if err != nil {
				return err
			}
			if err := sku.Create(&model.PmsProductSku{
				SpuID:          spuId,                        // 商品SpuId
				Name:           list.Name,                    // SKU名称
				SkuCode:        skuCode,                      // SKU编码
				MainPic:        list.MainPic,                 // 主图
				AlbumPics:      list.AlbumPics,               // 图片集
				Price:          float64(list.Price),          // 价格
				PromotionPrice: float64(list.PromotionPrice), // 单品促销价格
				Stock:          list.Stock,                   // 库存
				LowStock:       list.LowStock,                // 预警库存
				SpecData:       list.SpecData,                // 规格数据
				Weight:         float64(list.Weight),         // 重量(kg)
				PublishStatus:  list.PublishStatus,           // 上架状态：0-下架，1-上架
				VerifyStatus:   list.VerifyStatus,            // 审核状态：0-未审核，1-审核通过，2-审核不通过
				Sort:           list.Sort,                    // 排序
				CreateBy:       in.CreateBy,                  // 创建人ID
			}); err != nil {
				return err
			}
		}

		attr := qtx.PmsProductAttributeValue.WithContext(l.ctx)
		if _, err := attr.Where(qtx.PmsProductAttributeValue.SpuID.Eq(spuId)).Delete(); err != nil {
			return err
		}
		for _, list := range in.ProductAttributeValueList {
			if err := attr.Create(&model.PmsProductAttributeValue{
				SpuID:       spuId,                   // 商品SPU ID
				AttributeID: list.ProductAttributeId, // 属性ID
				Value:       list.AttributeValues,    // 属性值
				CreateBy:    in.CreateBy,             // 创建人ID
			}); err != nil {
				return err
			}
		}

		if err := logiccommon.RefreshSpuDraftSummary(l.ctx, tx, currentScope, spuId); err != nil {
			return err
		}

		return logiccommon.ApplyProductScope(l.ctx, tx, spuId, currentScope)
	})
	if err != nil {
		logc.Errorf(l.ctx, "更新商品SPU失败,参数:%+v,异常:%s", item, err.Error())
		return nil, err
	}

	sendProductESSync(l.ctx, l.svcCtx, spuId, currentScope)

	return &pmsclient.ProductSpuResp{}, nil
}
