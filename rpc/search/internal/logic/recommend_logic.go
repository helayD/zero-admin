package logic

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/bytedance/sonic"
	logiccommon "github.com/feihua/zero-admin/rpc/search/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/search/internal/svc"
	"github.com/feihua/zero-admin/rpc/search/search"

	"github.com/zeromicro/go-zero/core/logx"
)

type RecommendLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRecommendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecommendLogic {
	return &RecommendLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Recommend 根据商品id推荐商品
func (l *RecommendLogic) Recommend(in *search.RecommendReq) (*search.SearchResp, error) {
	current, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, errors.New("推荐范围无效")
	}

	pageNum, pageSize := logiccommon.NormalizePage(in.PageNum, in.PageSize)
	filters := logiccommon.ScopeFilters(current.ScopeType, current.PlatformID, current.TenantID, current.MerchantID)
	filters = append(filters, map[string]interface{}{
		"term": map[string]interface{}{"recommend_status": 1},
	})
	filters = logiccommon.AppendTermFilter(filters, "category_id", in.CategoryId)
	filters = logiccommon.AppendTermFilter(filters, "brand_id", in.BrandId)

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must":   logiccommon.KeywordMust(in.Keyword),
				"filter": filters,
			},
		},
		"from":             (pageNum - 1) * pageSize,
		"size":             pageSize,
		"track_total_hits": true,
		"sort": []interface{}{
			map[string]interface{}{"recommend_status_sort": map[string]string{"order": "desc"}},
			map[string]interface{}{"sales": map[string]string{"order": "desc"}},
		},
	}

	data, _ := sonic.Marshal(query)
	res, err := l.svcCtx.ESClient.Search(
		l.svcCtx.ESClient.Search.WithContext(l.ctx),
		l.svcCtx.ESClient.Search.WithIndex(svc.IndexName),
		l.svcCtx.ESClient.Search.WithBody(bytes.NewReader(data)),
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = res.Body.Close()
	}()
	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("es recommend error: %s", strings.TrimSpace(string(body)))
	}

	products, total, _, err := logiccommon.DecodeProductSearchResult(res.Body)
	if err != nil {
		return nil, err
	}

	return &search.SearchResp{
		Data:  products,
		Total: total,
	}, nil
}
