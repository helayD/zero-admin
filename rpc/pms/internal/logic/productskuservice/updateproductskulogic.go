package productskuservicelogic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/feihua/zero-admin/rpc/pms/gen/model"
	logiccommon "github.com/feihua/zero-admin/rpc/pms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// UpdateProductSkuLogic 更新商品SKU
/*
Author: LiuFeiHua
Date: 2025/06/16 14:37:37
*/
type UpdateProductSkuLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateProductSkuLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProductSkuLogic {
	return &UpdateProductSkuLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateProductSku 更新商品SKU
func (l *UpdateProductSkuLogic) UpdateProductSku(in *pmsclient.UpdateProductSkuReq) (*pmsclient.UpdateProductSkuResp, error) {
	if len(in.Data) == 0 {
		return nil, errors.New("缺少待更新的商品SKU")
	}

	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, in.Data[0].UpdateBy)
	if err != nil {
		return nil, err
	}

	var skuIds []int64
	var spuIDs []int64
	validationList := make([]logiccommon.ProductDraftSku, 0, len(in.Data))
	updates := make([]*model.PmsProductSku, 0, len(in.Data))
	for _, item := range in.Data {
		skuIds = append(skuIds, item.Id)
		spuIDs = append(spuIDs, item.SpuId)
		updateItem := &model.PmsProductSku{
			ID:             item.Id,
			SpuID:          item.SpuId,
			Name:           item.Name,
			SkuCode:        item.SkuCode,
			MainPic:        item.MainPic,
			AlbumPics:      item.AlbumPics,
			Price:          float64(item.Price),
			PromotionPrice: float64(item.PromotionPrice),
			Stock:          item.Stock,
			LowStock:       item.LowStock,
			SpecData:       item.SpecData,
			Weight:         float64(item.Weight),
			PublishStatus:  item.PublishStatus,
			VerifyStatus:   item.VerifyStatus,
			Sort:           item.Sort,
		}
		if len(item.PromotionStartTime) > 0 {
			startTime, _ := time.Parse("2006-01-02 15:04:05", item.PromotionStartTime)
			endTime, _ := time.Parse("2006-01-02 15:04:05", item.PromotionEndTime)
			updateItem.PromotionStartTime = &startTime
			updateItem.PromotionEndTime = &endTime
		}
		updates = append(updates, updateItem)
		validationList = append(validationList, logiccommon.ProductDraftSku{
			ID:             item.Id,
			SpuID:          item.SpuId,
			Name:           item.Name,
			SkuCode:        item.SkuCode,
			Price:          item.Price,
			PromotionPrice: item.PromotionPrice,
			Stock:          item.Stock,
			LowStock:       item.LowStock,
			SpecData:       item.SpecData,
		})
	}
	if _, err := logiccommon.EnsureSkuScope(l.ctx, l.svcCtx.DB, currentScope, skuIds, "pms.product_sku.update", in.Data[0].UpdateBy, "", "update sku"); err != nil {
		return nil, err
	}
	if _, err := logiccommon.EnsureProductScope(l.ctx, l.svcCtx.DB, currentScope, spuIDs, "pms.product_sku.update_parent", in.Data[0].UpdateBy, "", "update sku parent"); err != nil {
		return nil, errors.New("当前主体无权把SKU绑定到所选商品SPU")
	}

	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		spuTouched := make(map[int64]struct{})
		for idx, draft := range validationList {
			validatedSet, err := logiccommon.ComposeSkuValidationSet(l.ctx, tx, draft.SpuID, []logiccommon.ProductDraftSku{draft}, skuIds)
			if err != nil {
				return err
			}
			ignoreIDs := map[int64]struct{}{draft.ID: struct{}{}}
			generatedCode, err := logiccommon.EnsureSkuCode(l.ctx, tx, currentScope, draft.SpuID, draft.SkuCode, draft.SpecData, draft.Name, ignoreIDs)
			if err != nil {
				return fmt.Errorf("SKU编码冲突: %w", err)
			}
			validatedSet[len(validatedSet)-1].SkuCode = generatedCode
			if _, err := logiccommon.ValidateSkuDrafts(l.ctx, tx, currentScope, validatedSet, 0); err != nil {
				return err
			}
			updates[idx].SkuCode = generatedCode
			spuTouched[draft.SpuID] = struct{}{}
		}

		now := time.Now()
		for _, item := range updates {
			changes := map[string]interface{}{
				"spu_id":          item.SpuID,
				"name":            item.Name,
				"sku_code":        item.SkuCode,
				"main_pic":        item.MainPic,
				"album_pics":      item.AlbumPics,
				"price":           item.Price,
				"promotion_price": item.PromotionPrice,
				"promotion_start_time": item.PromotionStartTime,
				"promotion_end_time":   item.PromotionEndTime,
				"stock":                item.Stock,
				"low_stock":            item.LowStock,
				"spec_data":            item.SpecData,
				"weight":               item.Weight,
				"publish_status":       item.PublishStatus,
				"verify_status":        item.VerifyStatus,
				"sort":                 item.Sort,
				"update_by":            in.Data[0].UpdateBy,
				"update_time":          now,
			}
			if err := tx.WithContext(l.ctx).Table("pms_product_sku").Where("id = ?", item.ID).Updates(changes).Error; err != nil {
				return err
			}
			if err := logiccommon.ApplySkuScope(l.ctx, tx, item.ID, currentScope); err != nil {
				return err
			}
		}

		for spuID := range spuTouched {
			if err := logiccommon.RefreshSpuDraftSummary(l.ctx, tx, currentScope, spuID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		logc.Errorf(l.ctx, "更新sku的库存失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New(fmt.Sprintf("更新sku的库存失败,错误信息:%s", err.Error()))
	}

	return &pmsclient.UpdateProductSkuResp{}, nil
}
