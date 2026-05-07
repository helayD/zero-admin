package productspuservicelogic

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/pkg/pointerprocess"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/pkg/time_util"
	"github.com/feihua/zero-admin/rpc/pms/gen/model"
	logiccommon "github.com/feihua/zero-admin/rpc/pms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// QueryProductSpuListLogic 查询商品SPU列表
/*
Author: LiuFeiHua
Date: 2025/06/16 14:37:38
*/
type QueryProductSpuListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryProductSpuListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryProductSpuListLogic {
	return &QueryProductSpuListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// QueryProductSpuList 查询商品SPU列表
func (l *QueryProductSpuListLogic) QueryProductSpuList(in *pmsclient.QueryProductSpuListReq) (*pmsclient.QueryProductSpuListResp, error) {
	current, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		logc.Errorf(l.ctx, "查询商品SPU列表scope非法,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("查询商品SPU列表失败")
	}

	q := pkgscope.ApplyGovernanceScope(
		l.svcCtx.DB.WithContext(l.ctx).Model(&model.PmsProductSpu{}),
		current,
		"",
	)
	if len(in.Name) > 0 {
		q = q.Where("name LIKE ?", "%"+in.Name+"%")
	}
	if in.CategoryId != 0 {
		q = q.Where("category_id = ?", in.CategoryId)
	}
	if len(in.CategoryIds) > 0 {
		q = q.Where("category_ids LIKE ?", "%"+in.CategoryIds+"%")
	}
	if len(in.CategoryName) > 0 {
		q = q.Where("category_name LIKE ?", "%"+in.CategoryName+"%")
	}
	if in.BrandId != 0 {
		q = q.Where("brand_id = ?", in.BrandId)
	}
	if len(in.BrandName) > 0 {
		q = q.Where("brand_name LIKE ?", "%"+in.BrandName+"%")
	}

	if len(in.Keywords) > 0 {
		q = q.Where("keywords LIKE ?", "%"+in.Keywords+"%")
	}

	if in.PublishStatus != 2 {
		q = q.Where("publish_status = ?", in.PublishStatus)
	}
	if in.NewStatus != 2 {
		q = q.Where("new_status = ?", in.NewStatus)
	}
	if in.RecommendStatus != 2 {
		q = q.Where("recommend_status = ?", in.RecommendStatus)
	}
	if in.VerifyStatus != 2 {
		q = q.Where("verify_status = ?", in.VerifyStatus)
	}
	if in.PreviewStatus != 2 {
		q = q.Where("preview_status = ?", in.PreviewStatus)
	}

	if in.PromotionType != 6 {
		q = q.Where("promotion_type = ?", in.PromotionType)
	}

	var (
		result []model.PmsProductSpu
		count  int64
	)
	err = q.Session(&gorm.Session{}).Count(&count).Error
	if err == nil {
		err = q.Offset(int((in.PageNum - 1) * in.PageSize)).Limit(int(in.PageSize)).Find(&result).Error
	}

	if err != nil {
		logc.Errorf(l.ctx, "查询商品SPU列表失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("查询商品SPU列表失败")
	}

	reviewMetadata, err := loadProductReviewMetadata(l.ctx, l.svcCtx.ProductVertifyRecordModel, collectProductIDs(result))
	if err != nil {
		logc.Errorf(l.ctx, "查询商品审核元数据失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("查询商品SPU列表失败")
	}
	scopeMetadata, err := loadProductScopeMetadata(l.ctx, l.svcCtx.DB, current, collectProductIDs(result))
	if err != nil {
		logc.Errorf(l.ctx, "查询商品作用域元数据失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("查询商品SPU列表失败")
	}
	operationMetadata, err := loadProductOperationMetadata(l.ctx, l.svcCtx.DB, current, collectProductIDs(result))
	if err != nil {
		logc.Errorf(l.ctx, "查询商品状态操作元数据失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("查询商品SPU列表失败")
	}

	var list []*pmsclient.ProductSpuListData

	for _, item := range result {
		reviewMeta := reviewMetadata[item.ID]
		scopeMeta := scopeMetadata[item.ID]
		operationMeta := operationMetadata[item.ID]
		list = append(list, &pmsclient.ProductSpuListData{
			Id:                  item.ID,                                          // 商品SpuId
			ProductSn:           item.ProductSn,                                   // 商品货号
			Name:                item.Name,                                        // 商品名称
			CategoryId:          item.CategoryID,                                  // 商品分类ID
			CategoryIds:         item.CategoryIds,                                 // 商品分类ID集合
			CategoryName:        item.CategoryName,                                // 商品分类名称
			BrandId:             item.BrandID,                                     // 品牌ID
			BrandName:           item.BrandName,                                   // 品牌名称
			Unit:                item.Unit,                                        // 单位
			Weight:              float32(item.Weight),                             // 重量(kg)
			Keywords:            item.Keywords,                                    // 关键词
			AlbumPics:           item.AlbumPics,                                   // 画册图片，最多8张，以逗号分割
			MainPic:             item.MainPic,                                     // 主图
			PriceRange:          item.PriceRange,                                  // 价格区间
			PublishStatus:       item.PublishStatus,                               // 上架状态：0-下架，1-上架
			NewStatus:           item.NewStatus,                                   // 新品状态:0->不是新品；1->新品
			RecommendStatus:     item.RecommendStatus,                             // 推荐状态；0->不推荐；1->推荐
			VerifyStatus:        item.VerifyStatus,                                // 审核状态：0->未审核；1->审核通过
			PreviewStatus:       item.PreviewStatus,                               // 是否为预告商品：0->不是；1->是
			Sort:                item.Sort,                                        // 排序
			NewStatusSort:       item.NewStatusSort,                               // 新品排序
			RecommendStatusSort: item.RecommendStatusSort,                         // 推荐排序
			Sales:               item.Sales,                                       // 销量
			Stock:               item.Stock,                                       // 库存
			LowStock:            item.LowStock,                                    // 预警库存
		PromotionType:       item.PromotionType,                               // 促销类型：0->没有促销使用原价;1->使用促销价；2->使用会员价；3->使用阶梯价格；4->使用满减价格；5->秒杀
		FulfillmentMode:     item.FulfillmentMode,                             // 履约模式: physical_delivery-实物发货, digital_asset-数字资产
		FulfillmentRuleId:   item.FulfillmentRuleID,                           // 关联发卡规则ID,仅digital_asset模式时有效
		SubTitle:            item.SubTitle,                                    // 副标题
		DetailHtml:          item.DetailHTML,                                  // 产品详情网页内容
		DetailMobileHtml:    item.DetailMobileHTML,                            // 移动端网页详情
		CreateBy:            item.CreateBy,                                    // 创建人ID
		CreateTime:          time_util.TimeToStr(item.CreateTime),             // 创建时间
		UpdateBy:            pointerprocess.DefaltData(item.UpdateBy).(int64), // 更新人ID
		UpdateTime:          time_util.TimeToString(item.UpdateTime),          // 更新时间
		ScopeType:           scopeMeta.ScopeType,
		PlatformId:          scopeMeta.PlatformID,
		TenantId:            scopeMeta.TenantID,
		MerchantId:          scopeMeta.MerchantID,
		ReviewMan:           reviewMeta.ReviewMan,
		ReviewTime:          reviewMeta.ReviewTime,
		ReviewDetail:        reviewMeta.ReviewDetail,
		PublishMan:          operationMeta.PublishMan,
		PublishTime:         operationMeta.PublishTime,
		PublishDetail:       operationMeta.PublishDetail,
		RecommendMan:        operationMeta.RecommendMan,
		RecommendTime:       operationMeta.RecommendTime,
		RecommendDetail:     operationMeta.RecommendDetail,
	})
	}

	return &pmsclient.QueryProductSpuListResp{
		Total: count,
		List:  list,
	}, nil
}

func collectProductIDs(items []model.PmsProductSpu) []int64 {
	ids := make([]int64, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	return ids
}
