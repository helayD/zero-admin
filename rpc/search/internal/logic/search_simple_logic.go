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

type SearchSimpleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSearchSimpleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchSimpleLogic {
	return &SearchSimpleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SearchSimple 简单搜索-根据关键字通过名称或副标题查询商品
func (l *SearchSimpleLogic) SearchSimple(in *search.SearchSimpleReq) (*search.SearchResp, error) {
	current, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, errors.New("搜索范围无效")
	}

	pageNum, pageSize := logiccommon.NormalizePage(in.PageNum, in.PageSize)
	query := map[string]interface{}{
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
		"from":             (pageNum - 1) * pageSize,
		"size":             pageSize,
		"track_total_hits": true,
		"highlight": map[string]interface{}{
			"fields": map[string]interface{}{
				"name":  map[string]interface{}{},
				"brief": map[string]interface{}{},
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
		return nil, fmt.Errorf("es search error: %s", strings.TrimSpace(string(body)))
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
