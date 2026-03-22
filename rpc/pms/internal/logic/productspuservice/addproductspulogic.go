package productspuservicelogic

import (
	"context"
	"errors"
	"math/rand"
	"strconv"
	"time"

	"github.com/bytedance/sonic"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/pms/gen/model"
	"github.com/feihua/zero-admin/rpc/pms/gen/query"
	logiccommon "github.com/feihua/zero-admin/rpc/pms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// AddProductSpuLogic 添加商品SPU
/*
Author: LiuFeiHua
Date: 2025/06/16 14:37:38
*/
type AddProductSpuLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddProductSpuLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddProductSpuLogic {
	return &AddProductSpuLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AddProductSpu 添加商品SPU
// 逻辑步骤：
// 1.创建商品基本信息
// 2.会员价格
// 3.阶梯价格
// 4.满减价格
// 5.添加sku库存信息
// 6.添加商品参数,添加自定义商品规格
func (l *AddProductSpuLogic) AddProductSpu(in *pmsclient.ProductSpuReq) (*pmsclient.ProductSpuResp, error) {
	item := &model.PmsProductSpu{
		Name:                in.Name,                // 商品名称
		ProductSn:           in.ProductSn,           // 商品货号
		CategoryID:          in.CategoryId,          // 商品分类ID
		CategoryIds:         in.CategoryIds,         // 商品分类ID集合
		CategoryName:        in.CategoryName,        // 商品分类名称
		BrandID:             in.BrandId,             // 品牌ID
		BrandName:           in.BrandName,           // 品牌名称
		Unit:                in.Unit,                // 单位
		Weight:              float64(in.Weight),     // 重量(kg)
		Keywords:            in.Keywords,            // 关键词
		AlbumPics:           in.AlbumPics,           // 画册图片，最多8张，以逗号分割
		MainPic:             in.MainPic,             // 主图
		PriceRange:          in.PriceRange,          // 价格区间
		PublishStatus:       in.PublishStatus,       // 上架状态：0-下架，1-上架
		NewStatus:           in.NewStatus,           // 新品状态:0->不是新品；1->新品
		RecommendStatus:     in.RecommendStatus,     // 推荐状态；0->不推荐；1->推荐
		VerifyStatus:        in.VerifyStatus,        // 审核状态：0->未审核；1->审核通过
		PreviewStatus:       in.PreviewStatus,       // 是否为预告商品：0->不是；1->是
		Sort:                in.Sort,                // 排序
		NewStatusSort:       in.NewStatusSort,       // 新品排序
		RecommendStatusSort: in.RecommendStatusSort, // 推荐排序
		Sales:               in.Sales,               // 销量
		Stock:               in.Stock,               // 库存
		LowStock:            in.LowStock,            // 预警库存
		PromotionType:       in.PromotionType,       // 促销类型：0->没有促销使用原价;1->使用促销价；2->使用会员价；3->使用阶梯价格；4->使用满减价格；5->秒杀
		SubTitle:            in.SubTitle,            // 副标题
		DetailHTML:          in.DetailHtml,          // 产品详情网页内容
		DetailMobileHTML:    in.DetailMobileHtml,    // 移动端网页详情
		CreateBy:            in.CreateBy,            // 创建人ID
	}

	var (
		spuId        int64
		currentScope pkgscope.GovernanceScope
	)

	err := l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		if err := qtx.PmsProductSpu.WithContext(l.ctx).Create(item); err != nil {
			return err
		}

		spuId = item.ID

		memberPrice := qtx.PmsMemberPrice.WithContext(l.ctx)
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
		for _, list := range in.SkuStockList {
			skuCode := time.Now().Format("200601021504") + strconv.Itoa(rand.Intn(10))
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

		var err error
		currentScope, err = logiccommon.ResolveActorScope(l.ctx, tx, in.CreateBy)
		if err != nil {
			return err
		}

		return logiccommon.ApplyProductScope(l.ctx, tx, spuId, currentScope)
	})
	if err != nil {
		logc.Errorf(l.ctx, "添加商品SPU失败,参数:%+v,异常:%s", item, err.Error())
		return nil, errors.New("添加商品SPU失败")
	}

	body, _ := sonic.Marshal(pkgscope.NewProductESSyncPayload(spuId, currentScope))
	if err = l.svcCtx.RabbitMQ.SendMessage("product.event.exchange", "syn.product.to.es.queue", "syn.product.key", body); err != nil {
		logc.Errorf(l.ctx, "发送商品ES同步消息失败,spuId:%d,scope:%+v,异常:%s", spuId, currentScope, err.Error())
	}

	return &pmsclient.ProductSpuResp{
		SpuId: spuId,
	}, nil

}
