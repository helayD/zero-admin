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

type SearchRelatedLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSearchRelatedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchRelatedLogic {
	return &SearchRelatedLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SearchRelated 获取搜索的相关品牌、分类及筛选属性
func (l *SearchRelatedLogic) SearchRelated(in *search.SearchRelatedReq) (*search.SearchRelatedResp, error) {
	current, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, errors.New("搜索关联范围无效")
	}

	query := map[string]interface{}{
		"size": 0,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": logiccommon.KeywordMust(in.Keyword),
				"filter": logiccommon.ScopeFilters(
					current.ScopeType,
					current.PlatformID,
					current.TenantID,
					current.MerchantID,
				),
			},
		},
		"aggs": map[string]interface{}{
			"brand_names": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "brand_name.keyword",
					"size":  20,
				},
			},
			"category_names": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "category_name.keyword",
					"size":  20,
				},
			},
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
		return nil, fmt.Errorf("es search_related error: %s", strings.TrimSpace(string(body)))
	}

	_, _, aggregations, err := logiccommon.DecodeProductSearchResult(res.Body)
	if err != nil {
		return nil, err
	}

	return &search.SearchRelatedResp{
		BrandNames:    aggregations["brand_names"],
		CategoryNames: aggregations["category_names"],
	}, nil
}
