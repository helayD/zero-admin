package productskuservicelogic

import (
	"context"
	"fmt"
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

// AddProductSkuLogic 添加商品SKU
/*
Author: LiuFeiHua
Date: 2025/06/16 14:37:37
*/
type AddProductSkuLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddProductSkuLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddProductSkuLogic {
	return &AddProductSkuLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AddProductSku 添加商品SKU
func (l *AddProductSkuLogic) AddProductSku(in *pmsclient.AddProductSkuReq) (*pmsclient.AddProductSkuResp, error) {
	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, in.CreateBy)
	if err != nil {
		return nil, err
	}

	draft := logiccommon.ProductDraftSku{
		SpuID:          in.SpuId,
		Name:           in.Name,
		SkuCode:        in.SkuCode,
		Price:          in.Price,
		PromotionPrice: in.PromotionPrice,
		Stock:          in.Stock,
		LowStock:       in.LowStock,
		SpecData:       in.SpecData,
	}

	item := &model.PmsProductSku{}
	if len(in.PromotionStartTime) > 0 {
		startTime, _ := time.Parse("2006-01-02 15:04:05", in.PromotionStartTime)
		endTime, _ := time.Parse("2006-01-02 15:04:05", in.PromotionEndTime)
		item.PromotionStartTime = &startTime
		item.PromotionEndTime = &endTime
	}
	item.SpuID = in.SpuId
	item.MainPic = in.MainPic
	item.AlbumPics = in.AlbumPics
	item.Weight = float64(in.Weight)
	item.PublishStatus = in.PublishStatus
	item.VerifyStatus = in.VerifyStatus
	item.Sort = in.Sort
	item.CreateBy = in.CreateBy

	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		if err := logiccommon.EnsureSpuScopeForSku(l.ctx, tx, currentScope, in.SpuId, "pms.product_sku.add", in.CreateBy, "", fmt.Sprintf("spuId=%d", in.SpuId)); err != nil {
			return err
		}

		validatedSet, err := logiccommon.ComposeSkuValidationSet(l.ctx, tx, in.SpuId, []logiccommon.ProductDraftSku{draft}, nil)
		if err != nil {
			return err
		}
		generatedCode, err := logiccommon.EnsureSkuCode(l.ctx, tx, currentScope, in.SpuId, draft.SkuCode, draft.SpecData, draft.Name, nil)
		if err != nil {
			return fmt.Errorf("SKU编码冲突: %w", err)
		}
		validatedSet[len(validatedSet)-1].SkuCode = generatedCode
		if _, err := logiccommon.ValidateSkuDrafts(l.ctx, tx, currentScope, validatedSet, 0); err != nil {
			return err
		}

		item.Name = draft.Name
		item.SkuCode = generatedCode
		item.Price = float64(draft.Price)
		item.PromotionPrice = float64(draft.PromotionPrice)
		item.Stock = draft.Stock
		item.LowStock = draft.LowStock
		item.SpecData = draft.SpecData
		if err := qtx.PmsProductSku.WithContext(l.ctx).Create(item); err != nil {
			return err
		}
		if err := logiccommon.ApplySkuScope(l.ctx, tx, item.ID, currentScope); err != nil {
			return err
		}
		return logiccommon.RefreshSpuDraftSummary(l.ctx, tx, currentScope, in.SpuId)
	})
	if err != nil {
		logc.Errorf(l.ctx, "添加商品SKU失败,参数:%+v,异常:%s", item, err.Error())
		return nil, err
	}

	return &pmsclient.AddProductSkuResp{}, nil
}
