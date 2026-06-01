package common

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/feihua/zero-admin/rpc/search/search"
)

const (
	DefaultPageNum  int64 = 1
	DefaultPageSize int64 = 20
	MaxPageSize     int64 = 200
)

type ESTotal struct {
	Value int64 `json:"value"`
}

func (t *ESTotal) UnmarshalJSON(data []byte) error {
	raw := strings.TrimSpace(string(data))
	if raw == "" || raw == "null" {
		t.Value = 0
		return nil
	}

	if raw[0] == '{' {
		var payload struct {
			Value int64 `json:"value"`
		}
		if err := json.Unmarshal(data, &payload); err != nil {
			return err
		}
		t.Value = payload.Value
		return nil
	}

	return json.Unmarshal(data, &t.Value)
}

type ProductSearchResult struct {
	Hits struct {
		Total ESTotal `json:"total"`
		Hits  []struct {
			Source    search.ProductData  `json:"_source"`
			Highlight map[string][]string `json:"highlight"`
		} `json:"hits"`
	} `json:"hits"`
	Aggregations map[string]struct {
		Buckets []struct {
			Key string `json:"key"`
		} `json:"buckets"`
	} `json:"aggregations"`
}

func NormalizePage(pageNum, pageSize int64) (int64, int64) {
	if pageNum <= 0 {
		pageNum = DefaultPageNum
	}
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	return pageNum, pageSize
}

func ScopeFilters(scopeType string, platformID, tenantID, merchantID int64) []interface{} {
	filters := []interface{}{
		map[string]interface{}{"term": map[string]interface{}{"scope.scope_type": scopeType}},
	}

	if platformID > 0 {
		filters = append(filters, map[string]interface{}{"term": map[string]interface{}{"scope.platform_id": platformID}})
	}

	switch scopeType {
	case "tenant":
		if tenantID > 0 {
			filters = append(filters, map[string]interface{}{"term": map[string]interface{}{"scope.tenant_id": tenantID}})
		}
	case "merchant":
		if tenantID > 0 {
			filters = append(filters, map[string]interface{}{"term": map[string]interface{}{"scope.tenant_id": tenantID}})
		}
		if merchantID > 0 {
			filters = append(filters, map[string]interface{}{"term": map[string]interface{}{"scope.merchant_id": merchantID}})
		}
	}

	return filters
}

func AppendTermFilter(filters []interface{}, field string, value int64) []interface{} {
	if value <= 0 {
		return filters
	}

	return append(filters, map[string]interface{}{
		"term": map[string]interface{}{field: value},
	})
}

func KeywordMust(keyword string) []interface{} {
	trimmed := strings.TrimSpace(keyword)
	if trimmed == "" {
		return []interface{}{
			map[string]interface{}{"match_all": map[string]interface{}{}},
		}
	}

	return []interface{}{
		map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  trimmed,
				"fields": []string{"name", "brief", "keywords"},
			},
		},
	}
}

func DecodeProductSearchResult(body io.Reader) ([]*search.ProductData, int64, map[string][]string, error) {
	var result ProductSearchResult
	if err := json.NewDecoder(body).Decode(&result); err != nil {
		return nil, 0, nil, err
	}

	products := make([]*search.ProductData, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		product := hit.Source
		products = append(products, &product)
	}

	aggregations := make(map[string][]string, len(result.Aggregations))
	for name, agg := range result.Aggregations {
		values := make([]string, 0, len(agg.Buckets))
		for _, bucket := range agg.Buckets {
			if strings.TrimSpace(bucket.Key) == "" {
				continue
			}
			values = append(values, bucket.Key)
		}
		aggregations[name] = values
	}

	return products, result.Hits.Total.Value, aggregations, nil
}
