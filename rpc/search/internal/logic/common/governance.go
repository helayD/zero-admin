package common

import (
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/search/search"
)

func NormalizeProtoScope(input *search.GovernanceScope) (pkgscope.GovernanceScope, error) {
	if input == nil {
		return pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypePlatform, pkgscope.DefaultPlatformID, 0, 0)
	}

	return pkgscope.NormalizeGovernanceScope(input.ScopeType, input.PlatformId, input.TenantId, input.MerchantId)
}

func ProtoScope(current pkgscope.GovernanceScope) *search.GovernanceScope {
	return &search.GovernanceScope{
		ScopeType:  current.ScopeType,
		PlatformId: current.PlatformID,
		TenantId:   current.TenantID,
		MerchantId: current.MerchantID,
		ScopeLabel: current.Label(),
	}
}

func ProductDataToMap(p *search.ProductData) map[string]interface{} {
	m := map[string]interface{}{
		"id":                    p.Id,
		"name":                  p.Name,
		"category_id":          p.CategoryId,
		"category_ids":         p.CategoryIds,
		"category_name":        p.CategoryName,
		"brand_id":             p.BrandId,
		"brand_name":           p.BrandName,
		"unit":                 p.Unit,
		"weight":               p.Weight,
		"keywords":             p.Keywords,
		"brief":                p.Brief,
		"description":          p.Description,
		"album_pics":           p.AlbumPics,
		"main_pic":             p.MainPic,
		"price_range":          p.PriceRange,
		"publish_status":       p.PublishStatus,
		"new_status":           p.NewStatus,
		"recommend_status":     p.RecommendStatus,
		"verify_status":        p.VerifyStatus,
		"preview_status":       p.PreviewStatus,
		"sort":                 p.Sort,
		"new_status_sort":      p.NewStatusSort,
		"recommend_status_sort": p.RecommendStatusSort,
		"sales":                p.Sales,
		"stock":                p.Stock,
		"low_stock":            p.LowStock,
		"promotion_type":       p.PromotionType,
		"detail_title":         p.DetailTitle,
		"detail_desc":          p.DetailDesc,
		"detail_html":          p.DetailHtml,
		"detail_mobile_html":   p.DetailMobileHtml,
		"create_by":            p.CreateBy,
		"create_time":          p.CreateTime,
		"update_by":            p.UpdateBy,
		"update_time":          p.UpdateTime,
	}

	if p.Scope != nil {
		m["scope"] = map[string]interface{}{
			"scope_type":  p.Scope.ScopeType,
			"platform_id": p.Scope.PlatformId,
			"tenant_id":   p.Scope.TenantId,
			"merchant_id": p.Scope.MerchantId,
			"scope_label": p.Scope.ScopeLabel,
		}
	}

	return m
}
