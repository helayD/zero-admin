package search

import (
	"context"
	"strings"

	"github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/pkg/errorx"
	"github.com/feihua/zero-admin/rpc/search/search_client"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/status"
)

type SearchLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSearchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchLogic {
	return &SearchLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchLogic) Search(req *types.SearchReq) (resp *types.SearchResp, err error) {
	currentScope := common.ResolveEffectiveGovernanceScope(l.ctx)

	searchReq := &search_client.SearchReq{
		Keyword:    req.Keyword,
		PageNum:    req.PageNum,
		PageSize:   req.PageSize,
		Sort:       req.Sort,
		CategoryId: req.CategoryId,
		BrandId:    req.BrandId,
		Scope: &search_client.GovernanceScope{
			ScopeType:  currentScope.ScopeType,
			PlatformId: currentScope.PlatformID,
			TenantId:   currentScope.TenantID,
			MerchantId: currentScope.MerchantID,
			ScopeLabel: currentScope.Label(),
		},
	}

	searchResp, err := l.svcCtx.SearchClient.Search(l.ctx, searchReq)
	if err != nil {
		logc.Errorf(l.ctx, "搜索服务异常, 参数: %+v, 异常: %s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	var productItems []types.ProductItem
	for _, product := range searchResp.Data {
		priceStr := ""
		if product.PriceRange != "" {
			priceParts := strings.Split(product.PriceRange, "-")
			priceStr = priceParts[0]
		}
		productItems = append(productItems, types.ProductItem{
			Id:            product.Id,
			Name:          product.Name,
			Brief:         product.Brief,
			Price:         priceStr,
			OriginalPrice: 0,
			MainPic:       product.MainPic,
			Stock:         int(product.Stock),
			Sales:         int(product.Sales),
			CategoryId:    product.CategoryId,
			CategoryName:  product.CategoryName,
			BrandId:       product.BrandId,
			BrandName:     product.BrandName,
		})
	}

	isEmpty := searchResp.Total == 0
	var emptyHint string
	if isEmpty {
		emptyHint = "未找到符合条件的商品，请试试其他关键字或筛选条件"
	}

	return &types.SearchResp{
		Code:      0,
		Message:   "操作成功",
		Data:      productItems,
		Total:     searchResp.Total,
		Empty:     isEmpty,
		EmptyHint: emptyHint,
	}, nil
}
