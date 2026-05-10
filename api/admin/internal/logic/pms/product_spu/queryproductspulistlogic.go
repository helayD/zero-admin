package product_spu

import (
	"context"

	admincommon "github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"google.golang.org/grpc/status"

	"github.com/zeromicro/go-zero/core/logx"
)

// QueryProductSpuListLogic 查询商品SPU列表
/*
Author: LiuFeiHua
Date: 2025/06/16 14:37:38
*/
type QueryProductSpuListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryProductSpuListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryProductSpuListLogic {
	return &QueryProductSpuListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// QueryProductSpuList 查询商品SPU列表
func (l *QueryProductSpuListLogic) QueryProductSpuList(req *types.QueryProductSpuListReq) (resp *types.QueryProductSpuListResp, err error) {
	queryScope, err := admincommon.ResolveQueryGovernanceScope(l.ctx, admincommon.RequestedGovernanceScope{
		ScopeType:  req.ScopeType,
		PlatformID: req.PlatformId,
		TenantID:   req.TenantId,
		MerchantID: req.MerchantId,
	})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}

	// Story 10.10 Task 7: 商品列表按履约模式过滤。
	// 由于 pms-rpc 的 QueryProductSpuListReq 当前未携带 fulfillment_mode 字段，且其 .pb.go 是聚合文件
	// 改动会触发 25k+ 行重生成，本期采用 admin-api 层后过滤的折中：
	// 1) 当 fulfillmentMode 非空时，把 pms-rpc 的 PageSize 临时放大到一个较高上限 (1000)，禁用分页；
	// 2) admin-api 拉到列表后按 fulfillment_mode 过滤，再做内存分页。
	// 当商品总量超出 1000 时此实现会丢失尾部记录；后续 Story 升级 pms-rpc 协议后改为真正的 SQL 过滤。
	rpcReq := &pmsclient.QueryProductSpuListReq{
		PageNum:         req.Current,
		PageSize:        req.PageSize,
		Name:            req.Name,
		ProductSn:       req.ProductSn,
		CategoryId:      req.CategoryId,
		BrandId:         req.BrandId,
		Keywords:        req.Keywords,
		PublishStatus:   req.PublishStatus,
		NewStatus:       req.NewStatus,
		RecommendStatus: req.RecommendStatus,
		VerifyStatus:    req.VerifyStatus,
		PreviewStatus:   req.PreviewStatus,
		PromotionType:   req.PromotionType,
		Scope:           admincommon.PMSGovernanceScope(queryScope),
	}
	if req.FulfillmentMode != "" {
		rpcReq.PageNum = 1
		rpcReq.PageSize = 1000
	}

	result, err := l.svcCtx.ProductSpuService.QueryProductSpuList(l.ctx, rpcReq)

	if err != nil {
		logc.Errorf(l.ctx, "查询字商品SPU列表失败,参数：%+v,响应：%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	// [DEBUG-FULFILLMENT] 临时诊断日志：检查 RPC 返回字段是否真的丢失
	if len(result.List) > 0 {
		first := result.List[0]
		logc.Infof(l.ctx,
			"[DEBUG-FULFILLMENT] RPC 返回首条 id=%d name=%q fulfillmentMode=%q fulfillmentRuleId=%d total=%d",
			first.Id, first.Name, first.FulfillmentMode, first.FulfillmentRuleId, result.Total)
	}

	// Story 10.10 修复 H2: 折中方案在 total>1000 时会截断尾部记录，必须打告警日志方便 SRE/运营察觉。
	// 后续 Story 升级 pms-rpc 协议加 fulfillment_mode 字段后，可移除本段告警与上方 PageSize=1000 的临时放大。
	if req.FulfillmentMode != "" && result.Total > 1000 {
		logc.Errorf(l.ctx,
			"[ProductSpu履约模式过滤] 数据截断风险：fulfillmentMode=%s, scope=%+v, total=%d 超出折中阈值 1000，"+
				"过滤结果可能丢失尾部记录。请尽快完成 pms-rpc 协议升级以走 SQL 过滤。",
			req.FulfillmentMode, queryScope, result.Total)
	}

	// 按履约模式过滤
	filteredRows := result.List
	if req.FulfillmentMode != "" {
		filteredRows = filteredRows[:0]
		for _, item := range result.List {
			if item.FulfillmentMode == req.FulfillmentMode {
				filteredRows = append(filteredRows, item)
			}
		}
	}

	// 内存分页（仅当履约模式过滤生效时）
	totalAfterFilter := int64(len(filteredRows))
	pagedRows := filteredRows
	if req.FulfillmentMode != "" {
		page := req.Current
		if page <= 0 {
			page = 1
		}
		size := req.PageSize
		if size <= 0 {
			size = 20
		}
		start := int((page - 1) * size)
		end := start + int(size)
		if start >= len(filteredRows) {
			pagedRows = nil
		} else {
			if end > len(filteredRows) {
				end = len(filteredRows)
			}
			pagedRows = filteredRows[start:end]
		}
	}

	totalForResp := result.Total
	if req.FulfillmentMode != "" {
		totalForResp = totalAfterFilter
	}

	var list []*types.QueryProductSpuListData

	for _, detail := range pagedRows {
		list = append(list, &types.QueryProductSpuListData{
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
			FulfillmentMode:     detail.FulfillmentMode,     // 履约模式: physical_delivery-实物发货, digital_asset-数字资产
			FulfillmentRuleId:   detail.FulfillmentRuleId,   // 关联发卡规则ID,仅digital_asset模式时有效
			SubTitle:            detail.SubTitle,            // 详情标题
			DetailHtml:          detail.DetailHtml,          // 产品详情网页内容
			DetailMobileHtml:    detail.DetailMobileHtml,    // 移动端网页详情
			CreateBy:            detail.CreateBy,            // 创建人ID
			CreateTime:          detail.CreateTime,          // 创建时间
			UpdateBy:            detail.UpdateBy,            // 更新人ID
			UpdateTime:          detail.UpdateTime,          // 更新时间
			ScopeType:           detail.ScopeType,           // 作用域来源
			PlatformId:          detail.PlatformId,          // 平台ID
			TenantId:            detail.TenantId,            // 租户ID
			MerchantId:          detail.MerchantId,          // 商户ID
			ReviewMan:           detail.ReviewMan,           // 最近审核人
			ReviewTime:          detail.ReviewTime,          // 最近审核时间
			ReviewDetail:        detail.ReviewDetail,        // 最近审核意见
			PublishMan:          detail.PublishMan,          // 最近上下架操作人
			PublishTime:         detail.PublishTime,         // 最近上下架时间
			PublishDetail:       detail.PublishDetail,       // 最近上下架说明
			RecommendMan:        detail.RecommendMan,        // 最近推荐操作人
			RecommendTime:       detail.RecommendTime,       // 最近推荐时间
			RecommendDetail:     detail.RecommendDetail,     // 最近推荐说明
		})
	}

	return &types.QueryProductSpuListResp{
		Code:     "000000",
		Message:  "查询商品SPU列表成功",
		Data:     list,
		Current:  req.Current,
		PageSize: req.PageSize,
		Total:    totalForResp,
		Success:  true,
	}, nil
}
