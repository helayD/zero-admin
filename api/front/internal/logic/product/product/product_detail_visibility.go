package product

import (
	"strings"

	frontcommon "github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
)

const (
	productDetailCodeVisible     int64 = 0
	productDetailCodeHidden      int64 = 4101
	productDetailCodeNotSaleable int64 = 4102
)

func buildProductDetailResponse(
	detail *pmsclient.QueryProductSpuDetailResp,
	couponList *smsclient.QueryCouponByScopeIdResp,
	visibility frontcommon.ProductVisibilityResult,
) *types.QueryProductDetailResp {
	if !visibility.Visible {
		return &types.QueryProductDetailResp{
			Code:    productDetailCodeHidden,
			Message: visibility.ReasonMessage,
			Data:    buildUnavailableProductDetailData(visibility),
		}
	}

	code := productDetailCodeVisible
	message := "操作成功"
	if !visibility.Purchasable {
		code = productDetailCodeNotSaleable
		message = visibility.ReasonMessage
	}

	var couponData []*smsclient.CouponListData
	if couponList != nil {
		couponData = couponList.List
	}

	return &types.QueryProductDetailResp{
		Code:    code,
		Message: message,
		Data: types.ProductDetailData{
			ProductData:        buildProductData(detail),
			BrandData:          buildBrandData(detail),
			AttributeList:      buildProductAttributeListData(detail),
			AttributeValueList: buildProductAttributeValueListData(detail),
			SkuList:            buildSkuStockListData(detail),
			LadderList:         buildProductLadderListData(detail),
			FullList:           buildProductFullReductionListData(detail),
			MemberPriceList:    buildMemberPriceListData(detail),
			CouponList:         buildCouponListData(couponData),
			Visibility:         buildProductVisibilityData(visibility),
		},
	}
}

func buildUnavailableProductDetailData(visibility frontcommon.ProductVisibilityResult) types.ProductDetailData {
	return types.ProductDetailData{
		ProductData:        types.ProductData{},
		BrandData:          types.BrandData{},
		AttributeList:      []types.ProductAttributeList{},
		AttributeValueList: []types.ProductAttributeValueList{},
		SkuList:            []types.SkuStockList{},
		LadderList:         []types.ProductLadderList{},
		FullList:           []types.ProductFullReductionList{},
		MemberPriceList:    []types.MemberPriceList{},
		CouponList:         []types.CouponData{},
		Visibility:         buildProductVisibilityData(visibility),
	}
}

func buildProductVisibilityData(visibility frontcommon.ProductVisibilityResult) types.ProductVisibilityData {
	return types.ProductVisibilityData{
		Visible:        visibility.Visible,
		Purchasable:    visibility.Purchasable,
		ShowPrice:      visibility.ShowPrice,
		ShowStock:      visibility.ShowStock,
		Status:         visibility.Status,
		ReasonCode:     visibility.ReasonCode,
		ReasonMessage:  visibility.ReasonMessage,
		RecoveryHint:   visibility.RecoveryHint,
		FallbackAction: visibility.FallbackAction,
		FallbackTarget: visibility.FallbackTarget,
	}
}

func mapProductDetailQueryError(message string) (frontcommon.ProductVisibilityResult, bool) {
	switch {
	case strings.Contains(message, "商品SPU不存在"), strings.Contains(message, "商品不存在"):
		return frontcommon.BuildFrontProductVisibility(nil), true
	case strings.Contains(message, "租户已停用"), strings.Contains(message, "商户已停用"):
		return frontcommon.BuildOwnerDisabledVisibility(), true
	default:
		return frontcommon.ProductVisibilityResult{}, false
	}
}
