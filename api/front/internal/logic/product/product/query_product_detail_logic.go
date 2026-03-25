package product

import (
	"context"
	"encoding/json"
	"strings"

	frontcommon "github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/pkg/errorx"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"google.golang.org/grpc/status"

	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryProductDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryProductDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryProductDetailLogic {
	return &QueryProductDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// QueryProductDetail 获取商品详情
// 1.获取商品信息
// 2.获取品牌信息
// 3.获取商品属性信息
// 4.获取商品属性值信息
// 5.获取商品SKU库存信息
// 6.商品阶梯价格设置
// 7.商品满减价格设置
// 8.商品满减价格设置
// 9.商品可用优惠券
// 注意: 步骤1到7是在商品模块(rpc),9是在营销模块(rpc)
func (l *QueryProductDetailLogic) QueryProductDetail(req *types.QueryProductDetailReq) (resp *types.QueryProductDetailResp, err error) {
	currentScope := frontcommon.ResolveEffectiveGovernanceScope(l.ctx)

	detail, err := l.svcCtx.ProductSpuService.QueryProductSpuDetail(l.ctx, &pmsclient.QueryProductSpuDetailReq{
		Id:    req.ProductId,
		Scope: frontcommon.PMSGovernanceScope(currentScope),
	})

	if err != nil {
		logc.Errorf(l.ctx, "查询商品SPU详情失败,参数：%+v,响应：%s", req, err.Error())
		s, _ := status.FromError(err)
		if visibility, ok := mapProductDetailQueryError(s.Message()); ok {
			return buildProductDetailResponse(nil, nil, visibility, map[int64]int64{}), nil
		}
		return nil, errorx.NewDefaultError(s.Message())
	}

	visibility := frontcommon.BuildFrontProductVisibility(detail.Data)
	if !visibility.Visible {
		return buildProductDetailResponse(detail, nil, visibility, map[int64]int64{}), nil
	}

	// 8.商品可用优惠券(根据商品id和分类id查询)
	claimCountMap := map[int64]int64{}
	couponList, _ := l.svcCtx.CouponService.QueryCouponByScopeId(l.ctx, &smsclient.QueryCouponByScopeIdReq{
		ScopeIds: []int64{req.ProductId, detail.Data.CategoryId},
		Scope:    frontcommon.SMSGovernanceScope(currentScope),
	})
	if couponList != nil && len(couponList.List) > 0 {
		if memberId, ok := tryGetMemberID(l.ctx); ok {
			claimCountMap = l.buildMemberCouponClaimCount(memberId)
		}
	}

	return buildProductDetailResponse(detail, couponList, visibility, claimCountMap), nil
}

func (l *QueryProductDetailLogic) buildMemberCouponClaimCount(memberId int64) map[int64]int64 {
	claimCountMap := make(map[int64]int64)
	couponResp, err := l.svcCtx.CouponRecordService.QueryMemberCouponList(l.ctx, &smsclient.QueryMemberCouponListReq{
		MemberId: memberId,
		Status:   -1, // -1 表示查询所有状态的领取记录
	})
	if err != nil {
		logc.Errorf(l.ctx, "聚合会员优惠券领取记录失败,memberId:%d,异常:%s", memberId, err.Error())
		return claimCountMap
	}
	for _, item := range couponResp.List {
		claimCountMap[item.Id]++
	}
	return claimCountMap
}

func tryGetMemberID(ctx context.Context) (int64, bool) {
	rawMemberID := ctx.Value("memberId")
	if rawMemberID == nil {
		return 0, false
	}
	jsonNumber, ok := rawMemberID.(json.Number)
	if !ok {
		return 0, false
	}
	memberID, err := jsonNumber.Int64()
	if err != nil {
		return 0, false
	}
	return memberID, true
}

// 1.获取商品信息
func buildProductData(resp *pmsclient.QueryProductSpuDetailResp) types.ProductData {
	product := resp.Data
	price := strings.Split(product.PriceRange, "-")[0]
	return types.ProductData{
		Id:                  product.Id,                  // 商品SpuId
		Name:                product.Name,                // 商品名称
		ProductSn:           product.ProductSn,           // 商品货号
		CategoryId:          product.CategoryId,          // 商品分类ID
		CategoryIds:         product.CategoryIds,         // 商品分类ID集合
		CategoryName:        product.CategoryName,        // 商品分类名称
		BrandId:             product.BrandId,             // 品牌ID
		BrandName:           product.BrandName,           // 品牌名称
		Unit:                product.Unit,                // 单位
		Weight:              product.Weight,              // 重量(kg)
		Keywords:            product.Keywords,            // 关键词
		AlbumPics:           product.AlbumPics,           // 画册图片，最多8张，以逗号分割
		MainPic:             product.MainPic,             // 主图
		Price:               price,                       // 价格
		PriceRange:          product.PriceRange,          // 价格区间
		PublishStatus:       product.PublishStatus,       // 上架状态：0-下架，1-上架
		NewStatus:           product.NewStatus,           // 新品状态:0->不是新品；1->新品
		RecommendStatus:     product.RecommendStatus,     // 推荐状态；0->不推荐；1->推荐
		VerifyStatus:        product.VerifyStatus,        // 审核状态：0->未审核；1->审核通过
		PreviewStatus:       product.PreviewStatus,       // 是否为预告商品：0->不是；1->是
		Sort:                product.Sort,                // 排序
		NewStatusSort:       product.NewStatusSort,       // 新品排序
		RecommendStatusSort: product.RecommendStatusSort, // 推荐排序
		Sales:               product.Sales,               // 销量
		Stock:               product.Stock,               // 库存
		LowStock:            product.LowStock,            // 预警库存
		PromotionType:       product.PromotionType,       // 促销类型：0->没有促销使用原价;1->使用促销价；2->使用会员价；3->使用阶梯价格；4->使用满减价格；5->秒杀
		SubTitle:            product.SubTitle,            // 详情标题
		DetailHtml:          product.DetailHtml,          // 产品详情网页内容
		DetailMobileHtml:    product.DetailMobileHtml,    // 移动端网页详情
	}
}

// 2.获取品牌信息
func buildBrandData(resp *pmsclient.QueryProductSpuDetailResp) types.BrandData {
	detail := resp.Brand
	return types.BrandData{
		Id:                  detail.Id,                  //
		Name:                detail.Name,                // 品牌名称
		Logo:                detail.Logo,                // 品牌logo
		BigPic:              detail.BigPic,              // 专区大图
		Description:         detail.Description,         // 描述
		FirstLetter:         detail.FirstLetter,         // 首字母
		Sort:                detail.Sort,                // 排序
		RecommendStatus:     detail.RecommendStatus,     // 推荐状态
		ProductCount:        detail.ProductCount,        // 产品数量
		ProductCommentCount: detail.ProductCommentCount, // 产品评论数量
	}
}

// 3.获取商品属性值信息
func buildProductAttributeListData(resp *pmsclient.QueryProductSpuDetailResp) []types.ProductAttributeList {
	list := make([]types.ProductAttributeList, 0)
	for _, detail := range resp.ProductAttributeList {
		list = append(list, types.ProductAttributeList{
			Id:           detail.Id,           // 主键id
			GroupId:      detail.GroupId,      // 属性分组ID
			Name:         detail.Name,         // 属性名称
			InputType:    detail.InputType,    // 输入类型：1-手动输入，2-单选，3-多选
			ValueType:    detail.ValueType,    // 值类型：1-文本，2-数字，3-日期
			InputList:    detail.InputList,    // 可选值列表，用逗号分隔
			Unit:         detail.Unit,         // 单位
			IsRequired:   detail.IsRequired,   // 是否必填
			IsSearchable: detail.IsSearchable, // 是否支持搜索
			IsShow:       detail.IsShow,       // 是否显示
			Sort:         detail.Sort,         // 排序
		})
	}

	return list
}

// 4.获取商品属性值信息
func buildProductAttributeValueListData(resp *pmsclient.QueryProductSpuDetailResp) []types.ProductAttributeValueList {
	list := make([]types.ProductAttributeValueList, 0)
	for _, item := range resp.ProductAttributeValueList {
		list = append(list, types.ProductAttributeValueList{
			Id:          item.Id,
			SpuId:       item.SpuId,
			AttributeId: item.AttributeId,
			Value:       item.Value,
		})
	}

	return list
}

// 5.获取商品SKU库存信息
func buildSkuStockListData(resp *pmsclient.QueryProductSpuDetailResp) []types.SkuStockList {
	list := make([]types.SkuStockList, 0)
	for _, detail := range resp.SkuStockList {

		list = append(list, types.SkuStockList{
			Id:                 detail.Id,                 // 商品SpuId
			SpuId:              detail.SpuId,              // 商品SpuId
			Name:               detail.Name,               // SKU名称
			SkuCode:            detail.SkuCode,            // SKU编码
			MainPic:            detail.MainPic,            // 主图
			AlbumPics:          detail.AlbumPics,          // 图片集
			Price:              detail.Price,              // 价格
			PromotionPrice:     detail.PromotionPrice,     // 单品促销价格
			PromotionStartTime: detail.PromotionStartTime, // 促销开始时间
			PromotionEndTime:   detail.PromotionEndTime,   // 促销结束时间
			Stock:              detail.Stock,              // 库存
			LowStock:           detail.LowStock,           // 预警库存
			SpecData:           detail.SpecData,           // 规格数据
			Weight:             detail.Weight,             // 重量(kg)
			PublishStatus:      detail.PublishStatus,      // 上架状态：0-下架，1-上架
			VerifyStatus:       detail.VerifyStatus,       // 审核状态：0-未审核，1-审核通过，2-审核不通过
			Sort:               detail.Sort,               // 排序
			Sales:              detail.Sales,              // 销量

		})
	}

	return list
}

// 6.商品阶梯价格设置
func buildProductLadderListData(resp *pmsclient.QueryProductSpuDetailResp) []types.ProductLadderList {
	list := make([]types.ProductLadderList, 0)
	for _, detail := range resp.ProductLadderList {

		list = append(list, types.ProductLadderList{
			Id:        detail.Id,        //
			ProductId: detail.ProductId, // 商品id
			Count:     detail.Count,     // 满足的商品数量
			Discount:  detail.Discount,  // 折扣
			Price:     detail.Price,     // 折后价格
		})
	}

	return list

}

// 7.商品满减价格设置
func buildProductFullReductionListData(resp *pmsclient.QueryProductSpuDetailResp) []types.ProductFullReductionList {
	list := make([]types.ProductFullReductionList, 0)
	for _, detail := range resp.ProductFullReductionList {

		list = append(list, types.ProductFullReductionList{
			Id:          detail.Id,          //
			ProductId:   detail.ProductId,   // 商品id
			FullPrice:   detail.FullPrice,   // 商品满多少
			ReducePrice: detail.ReducePrice, // 商品减多少
		})
	}
	return list

}

// 8.会员价格
func buildMemberPriceListData(resp *pmsclient.QueryProductSpuDetailResp) []types.MemberPriceList {
	list := make([]types.MemberPriceList, 0)
	for _, detail := range resp.MemberPriceList {

		list = append(list, types.MemberPriceList{
			Id:              detail.Id,              //
			ProductId:       detail.ProductId,       // 商品id
			MemberLevelId:   detail.MemberLevelId,   // 会员等级id
			MemberPrice:     detail.MemberPrice,     // 会员价格
			MemberLevelName: detail.MemberLevelName, // 会员等级名称
		})
	}
	return list

}

// 9.商品优惠券
func buildCouponListData(resp []*smsclient.CouponListData, claimCountMap map[int64]int64) []types.CouponData {
	list := make([]types.CouponData, 0)
	for _, detail := range resp {
		// receiveStatus: 0=可领取, 1=已领取(达限领上限), 2=已领完, 3=未开始, 4=已过期
		receiveStatus := int32(0)
		if detail.PerLimit > 0 && claimCountMap[detail.Id] >= int64(detail.PerLimit) {
			receiveStatus = 1
		} else if detail.TotalCount > 0 && detail.ReceivedCount >= detail.TotalCount {
			receiveStatus = 2
		} else if detail.Status == 0 {
			receiveStatus = 3 // 未开始（草稿/待发布）
		} else if detail.Status == 2 || detail.Status == 3 {
			receiveStatus = 4 // 已过期/已取消
		}

		list = append(list, types.CouponData{
			Id:            detail.Id,          // 优惠券ID
			TypeId:        detail.TypeId,      // 优惠券类型ID
			Name:          detail.Name,        // 优惠券名称
			Code:          detail.Code,        // 优惠券码
			Amount:        detail.Amount,      // 优惠金额/折扣率
			MinAmount:     detail.MinAmount,   // 最低使用金额
			StartTime:     detail.StartTime,   // 生效时间
			EndTime:       detail.EndTime,     // 失效时间
			PerLimit:      detail.PerLimit,    // 每人限领数量
			Status:        detail.Status,      // 状态：0-未开始，1-进行中，2-已结束，3-已取消
			Description:   detail.Description, // 使用说明
			ReceiveStatus: receiveStatus,
			TotalCount:    detail.TotalCount,    // 发放总量
			ReceivedCount: detail.ReceivedCount, // 已领取数量
		})
	}
	return list

}
