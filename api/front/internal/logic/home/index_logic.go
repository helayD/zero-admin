package home

import (
	"context"
	"strings"
	"time"

	frontcommon "github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/cms/cmsclient"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"

	"github.com/zeromicro/go-zero/core/logx"
)

// IndexLogic 获取首页数据
/*
Author: LiuFeiHua
Date: 2025/6/19 15:43
*/
type IndexLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewIndexLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IndexLogic {
	return &IndexLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *IndexLogic) Index(req *types.HomeReq) (resp *types.HomeResp, err error) {
	currentScope := frontcommon.ResolveEffectiveGovernanceScope(l.ctx)

	return &types.HomeResp{
		Code:    0,
		Message: "操作成功",
		Data: types.Data{
			AdvertiseList:      queryAdvertiseList(l, currentScope),
			BrandList:          queryBrandList(l, req, currentScope),
			HomeFlashPromotion: queryHomeFlashPromotion(l, req),
			NewProductList:     queryNewProductList(l, req, currentScope),
			HotProductList:     queryHotProductList(l, req, currentScope),
			SubjectList:        querySubjectList(l, req, currentScope),       // 推荐专题
			PreferredAreaList:  queryPreferredAreaList(l, req, currentScope), // 优选专区
		},
	}, nil
}

// 推荐专题
func querySubjectList(l *IndexLogic, req *types.HomeReq, currentScope pkgscope.GovernanceScope) []types.SubjectList {
	var list []types.SubjectList
	res, err := l.svcCtx.SubjectService.QuerySubjectList(l.ctx, &cmsclient.QuerySubjectListReq{
		PageNum:         1,
		PageSize:        int64(req.SubjectNumber),
		Title:           "", // 专题标题
		RecommendStatus: 1,  // 推荐状态：0->不推荐；1->推荐
		ShowStatus:      1,  // 显示状态：0->不显示；1->显示
		Scope:           frontcommon.CMSGovernanceScope(currentScope),
	})

	if err != nil || res == nil {
		l.Errorf("querySubjectList failed: req=%+v scope=%+v err=%v", req, currentScope, err)
		return list
	}

	for _, item := range res.List {
		list = append(list, types.SubjectList{
			Id:              item.Id,              // 专题id
			CategoryId:      item.CategoryId,      // 专题分类id
			Title:           item.Title,           // 专题标题
			Pic:             item.Pic,             // 专题主图
			ProductCount:    item.ProductCount,    // 关联产品数量
			RecommendStatus: item.RecommendStatus, // 推荐状态：0->不推荐；1->推荐
			CollectCount:    item.CollectCount,    // 收藏数
			ReadCount:       item.ReadCount,       // 阅读数
			CommentCount:    item.CommentCount,    // 评论数
			AlbumPics:       item.AlbumPics,       // 画册图片用逗号分割
			Description:     item.Description,     // 专题内容
			ShowStatus:      item.ShowStatus,      // 显示状态：0->不显示；1->显示
			Content:         item.Content,         // 专题内容
			ForwardCount:    item.ForwardCount,    // 转发数
			CategoryName:    item.CategoryName,    // 专题分类名称
			Sort:            1,
		})
	}
	return list
}

// 优选专区
func queryPreferredAreaList(l *IndexLogic, req *types.HomeReq, currentScope pkgscope.GovernanceScope) []types.PreferredAreaListData {
	var list []types.PreferredAreaListData
	res, err := l.svcCtx.PreferredAreaService.QueryPreferredAreaList(l.ctx, &cmsclient.QueryPreferredAreaListReq{
		PageNum:    1,
		PageSize:   int64(req.PreferredAreaNumber),
		ShowStatus: 1, // 显示状态：1->显示
		Scope:      frontcommon.CMSGovernanceScope(currentScope),
	})

	if err != nil || res == nil {
		l.Errorf("queryPreferredAreaList failed: req=%+v scope=%+v err=%v", req, currentScope, err)
		return list
	}

	for _, item := range res.List {
		list = append(list, types.PreferredAreaListData{
			Id:         item.Id,         // 专区id
			Name:       item.Name,       // 专区名称
			SubTitle:   item.SubTitle,   // 子标题
			Pic:        item.Pic,        // 展示图片
			Sort:       item.Sort,       // 排序
			ShowStatus: item.ShowStatus, // 显示状态
		})
	}
	return list
}

// 人气推荐
func queryHotProductList(l *IndexLogic, req *types.HomeReq, currentScope pkgscope.GovernanceScope) []types.IndexProductData {
	resp, err := l.svcCtx.ProductSpuService.QueryProductSpuList(l.ctx, &pmsclient.QueryProductSpuListReq{
		PageNum:         1,
		PageSize:        req.HotProductNumber,
		Name:            "",
		CategoryId:      0, // 商品分类ID
		BrandId:         0, // 品牌ID
		PublishStatus:   1, // 上架状态：0-下架，1-上架
		NewStatus:       2, // 新品状态:0->不是新品；1->新品
		RecommendStatus: 1, // 推荐状态；0->不推荐；1->推荐
		VerifyStatus:    1, // 审核状态：0->未审核；1->审核通过
		PreviewStatus:   0, // 是否为预告商品：0->不是；1->是
		PromotionType:   6, // 促销类型：0->没有促销使用原价;1->使用促销价；2->使用会员价；3->使用阶梯价格；4->使用满减价格；5->秒杀
		Scope:           frontcommon.PMSGovernanceScope(currentScope),
	})

	var list []types.IndexProductData
	if err != nil || resp == nil {
		l.Errorf("queryHotProductList failed: req=%+v scope=%+v err=%v", req, currentScope, err)
		return list
	}

	for _, detail := range resp.List {
		if err := frontcommon.EnsureFrontProductVisible(detail); err != nil {
			continue
		}
		price := strings.Split(detail.PriceRange, "-")[0]
		list = append(list, types.IndexProductData{
			Id:                  detail.Id,                  // 商品SpuId
			Name:                detail.Name,                // 商品名称
			ProductSn:           detail.ProductSn,           // 商品货号
			CategoryId:          detail.CategoryId,          // 商品分类ID
			CategoryIds:         detail.CategoryIds,         // 商品分类ID集合
			CategoryName:        detail.CategoryName,        // 商品分类名称
			BrandId:             detail.BrandId,             // 品牌ID
			BrandName:           detail.BrandName,           // 品牌名称
			Unit:                detail.Unit,                // 单位
			Weight:              detail.Weight,              // 重量(kg)
			Keywords:            detail.Keywords,            // 关键词
			AlbumPics:           detail.AlbumPics,           // 画册图片，最多8张，以逗号分割
			MainPic:             detail.MainPic,             // 主图
			Price:               price,                      // 价格
			PriceRange:          detail.PriceRange,          // 价格区间
			PublishStatus:       detail.PublishStatus,       // 上架状态：0-下架，1-上架
			NewStatus:           detail.NewStatus,           // 新品状态:0->不是新品；1->新品
			RecommendStatus:     detail.RecommendStatus,     // 推荐状态；0->不推荐；1->推荐
			VerifyStatus:        detail.VerifyStatus,        // 审核状态：0->未审核；1->审核通过
			PreviewStatus:       detail.PreviewStatus,       // 是否为预告商品：0->不是；1->是
			Sort:                detail.Sort,                // 排序
			NewStatusSort:       detail.NewStatusSort,       // 新品排序
			RecommendStatusSort: detail.RecommendStatusSort, // 推荐排序
			Sales:               detail.Sales,               // 销量
			Stock:               detail.Stock,               // 库存
			LowStock:            detail.LowStock,            // 预警库存
			PromotionType:       detail.PromotionType,       // 促销类型：0->没有促销使用原价;1->使用促销价；2->使用会员价；3->使用阶梯价格；4->使用满减价格；5->秒杀
			SubTitle:            detail.SubTitle,            // 详情标题
			DetailHtml:          detail.DetailHtml,          // 产品详情网页内容
			DetailMobileHtml:    detail.DetailMobileHtml,    // 移动端网页详情
		})
	}
	return list
}

// 新品推荐
func queryNewProductList(l *IndexLogic, req *types.HomeReq, currentScope pkgscope.GovernanceScope) []types.IndexProductData {
	resp, err := l.svcCtx.ProductSpuService.QueryProductSpuList(l.ctx, &pmsclient.QueryProductSpuListReq{
		PageNum:         1,
		PageSize:        req.NewProductNumber,
		CategoryId:      0, // 商品分类ID
		BrandId:         0, // 品牌ID
		PublishStatus:   1, // 上架状态：0-下架，1-上架
		NewStatus:       1, // 新品状态:0->不是新品；1->新品
		RecommendStatus: 2, // 推荐状态；0->不推荐；1->推荐
		VerifyStatus:    1, // 审核状态：0->未审核；1->审核通过
		PreviewStatus:   0, // 是否为预告商品：0->不是；1->是
		PromotionType:   6, // 促销类型：0->没有促销使用原价;1->使用促销价；2->使用会员价；3->使用阶梯价格；4->使用满减价格；5->秒杀
		Scope:           frontcommon.PMSGovernanceScope(currentScope),
	})

	var list []types.IndexProductData
	if err != nil || resp == nil {
		l.Errorf("queryNewProductList failed: req=%+v scope=%+v err=%v", req, currentScope, err)
		return list
	}

	for _, detail := range resp.List {
		if err := frontcommon.EnsureFrontProductVisible(detail); err != nil {
			continue
		}
		price := strings.Split(detail.PriceRange, "-")[0]
		list = append(list, types.IndexProductData{
			Id:                  detail.Id,                  // 商品SpuId
			Name:                detail.Name,                // 商品名称
			ProductSn:           detail.ProductSn,           // 商品货号
			CategoryId:          detail.CategoryId,          // 商品分类ID
			CategoryIds:         detail.CategoryIds,         // 商品分类ID集合
			CategoryName:        detail.CategoryName,        // 商品分类名称
			BrandId:             detail.BrandId,             // 品牌ID
			BrandName:           detail.BrandName,           // 品牌名称
			Unit:                detail.Unit,                // 单位
			Weight:              detail.Weight,              // 重量(kg)
			Keywords:            detail.Keywords,            // 关键词
			AlbumPics:           detail.AlbumPics,           // 画册图片，最多8张，以逗号分割
			MainPic:             detail.MainPic,             // 主图
			Price:               price,                      // 价格
			PriceRange:          detail.PriceRange,          // 价格区间
			PublishStatus:       detail.PublishStatus,       // 上架状态：0-下架，1-上架
			NewStatus:           detail.NewStatus,           // 新品状态:0->不是新品；1->新品
			RecommendStatus:     detail.RecommendStatus,     // 推荐状态；0->不推荐；1->推荐
			VerifyStatus:        detail.VerifyStatus,        // 审核状态：0->未审核；1->审核通过
			PreviewStatus:       detail.PreviewStatus,       // 是否为预告商品：0->不是；1->是
			Sort:                detail.Sort,                // 排序
			NewStatusSort:       detail.NewStatusSort,       // 新品排序
			RecommendStatusSort: detail.RecommendStatusSort, // 推荐排序
			Sales:               detail.Sales,               // 销量
			Stock:               detail.Stock,               // 库存
			LowStock:            detail.LowStock,            // 预警库存
			PromotionType:       detail.PromotionType,       // 促销类型：0->没有促销使用原价;1->使用促销价；2->使用会员价；3->使用阶梯价格；4->使用满减价格；5->秒杀
			SubTitle:            detail.SubTitle,            // 详情标题
			DetailHtml:          detail.DetailHtml,          // 产品详情网页内容
			DetailMobileHtml:    detail.DetailMobileHtml,    // 移动端网页详情
		})
	}
	return list
}

// 当前秒杀场次
func queryHomeFlashPromotion(l *IndexLogic, req *types.HomeReq) types.HomeFlashPromotion {
	var resp types.HomeFlashPromotion
	currentScope := frontcommon.ResolveEffectiveGovernanceScope(l.ctx)

	// 1. 查询当天是否有秒杀活动
	currentDate := time.Now().Format("2006-01-02")
	activityResp, err := l.svcCtx.SeckillActivityService.QuerySeckillActivityListByDate(l.ctx, &smsclient.QuerySeckillActivityListByDateReq{
		CurrentDate: currentDate,
	})
	if err != nil || activityResp == nil || len(activityResp.List) == 0 {
		// 无秒杀活动时降级展示新品推荐
		resp.ProductList = queryNewProductList(l, req, currentScope)
		return resp
	}

	// 2. 查询当前时间是否有秒杀场次
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	sessionResp, err := l.svcCtx.SeckillSessionService.QuerySeckillSessionListByTime(l.ctx, &smsclient.QuerySeckillSessionListByTimeReq{
		CurrentTIme: currentTime,
	})
	if err != nil || sessionResp == nil || len(sessionResp.List) == 0 {
		// 无当前场次时降级展示新品推荐
		resp.ProductList = queryNewProductList(l, req, currentScope)
		return resp
	}

	currentSession := sessionResp.List[0]
	resp.StartTime = currentSession.StartTime
	resp.EndTime = currentSession.EndTime

	// 查询下一场次时间
	nextSessionResp, _ := l.svcCtx.SeckillSessionService.QuerySeckillSessionListByTime(l.ctx, &smsclient.QuerySeckillSessionListByTimeReq{
		CurrentTIme: currentSession.EndTime,
	})
	if nextSessionResp != nil && len(nextSessionResp.List) > 0 {
		nextSession := nextSessionResp.List[0]
		resp.NextStartTime = nextSession.StartTime
		resp.NextEndTime = nextSession.EndTime
	}

	// 3. 查询秒杀商品列表
	seckillProductResp, err := l.svcCtx.SeckillProductService.QuerySeckillProductList(l.ctx, &smsclient.QuerySeckillProductListReq{
		ActivityId: activityResp.List[0].Id,
		SessionId:  currentSession.Id,
		Status:     1,
		PageNum:    1,
		PageSize:   100,
	})
	if err != nil || seckillProductResp == nil || len(seckillProductResp.List) == 0 {
		resp.ProductList = queryNewProductList(l, req, currentScope)
		return resp
	}

	// 4. 通过 SKU ID 查询对应的 SPU ID，再批量查询 SPU 商品详情
	// TODO(perf): 此处为 N+1 查询——每个秒杀商品单独调一次 QueryProductSkuDetail RPC。
	// 应在 pms-rpc 新增 QueryProductSkuListByIds 批量接口后改为一次批量调用。
	spuIdSet := make(map[int64]bool)
	for _, item := range seckillProductResp.List {
		skuDetail, skuErr := l.svcCtx.ProductSkuService.QueryProductSkuDetail(l.ctx, &pmsclient.QueryProductSkuDetailReq{
			Id: item.SkuId,
		})
		if skuErr != nil || skuDetail == nil {
			continue
		}
		if skuDetail.SpuId > 0 {
			spuIdSet[skuDetail.SpuId] = true
		}
	}

	var spuIds []int64
	for id := range spuIdSet {
		spuIds = append(spuIds, id)
	}

	if len(spuIds) == 0 {
		resp.ProductList = queryNewProductList(l, req, currentScope)
		return resp
	}

	spuResp, err := l.svcCtx.ProductSpuService.QueryProductSpuListByIds(l.ctx, &pmsclient.QueryProductSpuByIdsReq{
		Ids:   spuIds,
		Scope: frontcommon.PMSGovernanceScope(currentScope),
	})
	if err != nil || spuResp == nil || len(spuResp.List) == 0 {
		resp.ProductList = queryNewProductList(l, req, currentScope)
		return resp
	}

	var productList []types.IndexProductData
	for _, detail := range spuResp.List {
		if err := frontcommon.EnsureFrontProductVisible(detail); err != nil {
			continue
		}
		price := strings.Split(detail.PriceRange, "-")[0]
		productList = append(productList, types.IndexProductData{
			Id:                  detail.Id,
			Name:                detail.Name,
			ProductSn:           detail.ProductSn,
			CategoryId:          detail.CategoryId,
			CategoryIds:         detail.CategoryIds,
			CategoryName:        detail.CategoryName,
			BrandId:             detail.BrandId,
			BrandName:           detail.BrandName,
			Unit:                detail.Unit,
			Weight:              detail.Weight,
			Keywords:            detail.Keywords,
			AlbumPics:           detail.AlbumPics,
			MainPic:             detail.MainPic,
			Price:               price,
			PriceRange:          detail.PriceRange,
			PublishStatus:       detail.PublishStatus,
			NewStatus:           detail.NewStatus,
			RecommendStatus:     detail.RecommendStatus,
			VerifyStatus:        detail.VerifyStatus,
			PreviewStatus:       detail.PreviewStatus,
			Sort:                detail.Sort,
			NewStatusSort:       detail.NewStatusSort,
			RecommendStatusSort: detail.RecommendStatusSort,
			Sales:               detail.Sales,
			Stock:               detail.Stock,
			LowStock:            detail.LowStock,
			PromotionType:       5, // 秒杀
			SubTitle:            detail.SubTitle,
			DetailHtml:          detail.DetailHtml,
			DetailMobileHtml:    detail.DetailMobileHtml,
		})
	}
	resp.ProductList = productList
	return resp
}

// 推荐品牌
func queryBrandList(l *IndexLogic, req *types.HomeReq, currentScope pkgscope.GovernanceScope) []types.IndexBrandData {
	result, err := l.svcCtx.ProductBrandService.QueryProductBrandList(l.ctx, &pmsclient.QueryProductBrandListReq{
		PageNum:         1,
		PageSize:        req.BrandNumber,
		Name:            "", // 品牌名称
		RecommendStatus: 1,  // 推荐状态
		IsEnabled:       1,  // 是否启用
		Scope:           frontcommon.PMSGovernanceScope(currentScope),
	})

	var list []types.IndexBrandData
	if err != nil || result == nil {
		l.Errorf("queryBrandList failed: req=%+v err=%v", req, err)
		return list
	}

	for _, detail := range result.List {
		list = append(list, types.IndexBrandData{
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

		})
	}
	return list
}

// 获取轮播广告
// 注意：RPC 的 StartTime/EndTime 参数语义是管理后台范围搜索（start_time >= value, end_time <= value），
// 与前台需要的“当前有效广告”过滤方向相反，因此在 front-api 侧做客户端时间过滤。
func queryAdvertiseList(l *IndexLogic, currentScope pkgscope.GovernanceScope) []types.AdvertiseList {
	now := time.Now()
	_ = currentScope // 广告位 RPC 暂无 scope 字段，预留作用域参数供后续扩展
	result, err := l.svcCtx.HomeAdvertiseService.QueryHomeAdvertiseList(l.ctx, &smsclient.QueryHomeAdvertiseListReq{
		PageNum:  1,
		PageSize: 100,
		Type:     1, // 轮播位置：0->PC首页轮播；1->app首页轮播
		Status:   1, // 上下线状态：0->下线；1->上线
	})

	var list []types.AdvertiseList
	if err != nil || result == nil {
		l.Errorf("queryAdvertiseList failed: pageNum=%d pageSize=%d type=%d status=%d err=%v", 1, 100, 1, 1, err)
		return list
	}

	for _, detail := range result.List {
		// 客户端时间过滤：只保留当前时间在 [start_time, end_time] 区间内的广告
		if len(detail.StartTime) > 0 {
			if startTime, err := time.Parse("2006-01-02 15:04:05", detail.StartTime); err == nil && now.Before(startTime) {
				continue // 广告尚未生效
			}
		}
		if len(detail.EndTime) > 0 {
			if endTime, err := time.Parse("2006-01-02 15:04:05", detail.EndTime); err == nil && now.After(endTime) {
				continue // 广告已过期
			}
		}
		list = append(list, types.AdvertiseList{
			Id:         detail.Id,         // 编号
			Name:       detail.Name,       // 名称
			Type:       detail.Type,       // 轮播位置：0->PC首页轮播；1->app首页轮播
			Pic:        detail.Pic,        // 图片地址
			StartTime:  detail.StartTime,  // 开始时间
			EndTime:    detail.EndTime,    // 结束时间
			Status:     detail.Status,     // 上下线状态：0->下线；1->上线
			ClickCount: detail.ClickCount, // 点击数
			OrderCount: detail.OrderCount, // 下单数
			Url:        detail.Url,        // 链接地址
			Remark:     detail.Remark,     // 备注
			Sort:       detail.Sort,       // 排序

		})
	}

	return list
}
