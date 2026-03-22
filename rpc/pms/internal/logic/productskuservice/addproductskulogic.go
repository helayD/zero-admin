package productskuservicelogic

import (
	"context"
	"fmt"
	"github.com/feihua/zero-admin/rpc/pms/gen/model"
	"github.com/feihua/zero-admin/rpc/pms/gen/query"
	logiccommon "github.com/feihua/zero-admin/rpc/pms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"time"
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

	item := &model.PmsProductSku{
		SpuID:          in.SpuId,                   // 商品SpuId
		Name:           in.Name,                    // SKU名称
		SkuCode:        in.SkuCode,                 // SKU编码
		MainPic:        in.MainPic,                 // 主图
		AlbumPics:      in.AlbumPics,               // 图片集
		Price:          float64(in.Price),          // 价格
		PromotionPrice: float64(in.PromotionPrice), // 单品促销价格
		Stock:          in.Stock,                   // 库存
		LowStock:       in.LowStock,                // 预警库存
		SpecData:       in.SpecData,                // 规格数据
		Weight:         float64(in.Weight),         // 重量(kg)
		PublishStatus:  in.PublishStatus,           // 上架状态：0-下架，1-上架
		VerifyStatus:   in.VerifyStatus,            // 审核状态：0-未审核，1-审核通过，2-审核不通过
		Sort:           in.Sort,                    // 排序
		CreateBy:       in.CreateBy,                // 创建人ID
	}

	if len(in.PromotionStartTime) > 0 {
		startTime, _ := time.Parse("2006-01-02 15:04:05", in.PromotionStartTime)
		endTime, _ := time.Parse("2006-01-02 15:04:05", in.PromotionEndTime)
		item.PromotionStartTime = &startTime // 促销开始时间
		item.PromotionEndTime = &endTime     // 促销结束时间
	}
	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		if err := logiccommon.EnsureSpuScopeForSku(l.ctx, tx, currentScope, in.SpuId, "pms.product_sku.add", in.CreateBy, "", fmt.Sprintf("spuId=%d", in.SpuId)); err != nil {
			return err
		}
		if err := qtx.PmsProductSku.WithContext(l.ctx).Create(item); err != nil {
			return err
		}
		return logiccommon.ApplySkuScope(l.ctx, tx, item.ID, currentScope)
	})
	if err != nil {
		logc.Errorf(l.ctx, "添加商品SKU失败,参数:%+v,异常:%s", item, err.Error())
		return nil, err
	}

	return &pmsclient.AddProductSkuResp{}, nil
}
