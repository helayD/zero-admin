package common

import (
	"errors"

	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
)

const (
	FrontProductVisibilityStatusVisible     = "visible"
	FrontProductVisibilityStatusHidden      = "hidden"
	FrontProductVisibilityStatusNotSaleable = "not-saleable"

	FrontProductVisibilityReasonNone          = ""
	FrontProductVisibilityReasonNotFound      = "product_not_found"
	FrontProductVisibilityReasonOwnerDisabled = "owner_disabled"
	FrontProductVisibilityReasonPendingReview = "pending_review"
	FrontProductVisibilityReasonRejected      = "review_rejected"
	FrontProductVisibilityReasonOffShelf      = "off_shelf"
	FrontProductVisibilityReasonPreviewOnly   = "preview_only"
	FrontProductVisibilityReasonSoldOut       = "sold_out"

	FrontProductFallbackBrowseProductList = "browse_product_list"
	FrontProductFallbackBrowseSimilar     = "browse_similar"
	FrontProductFallbackGoHome            = "go_home"
)

type ProductVisibilityResult struct {
	Visible        bool
	Purchasable    bool
	ShowPrice      bool
	ShowStock      bool
	Status         string
	ReasonCode     string
	ReasonMessage  string
	RecoveryHint   string
	FallbackAction string
	FallbackTarget string
}

func BuildFrontProductVisibility(product *pmsclient.ProductSpuListData) ProductVisibilityResult {
	if product == nil {
		return ProductVisibilityResult{
			Visible:        false,
			Purchasable:    false,
			ShowPrice:      false,
			ShowStock:      false,
			Status:         FrontProductVisibilityStatusHidden,
			ReasonCode:     FrontProductVisibilityReasonNotFound,
			ReasonMessage:  "商品不存在或当前作用域不可见",
			RecoveryHint:   "请返回商品列表重新选择在售商品",
			FallbackAction: FrontProductFallbackBrowseProductList,
			FallbackTarget: "product_list",
		}
	}

	switch product.VerifyStatus {
	case 1:
	case 2:
		return ProductVisibilityResult{
			Visible:        false,
			Purchasable:    false,
			ShowPrice:      false,
			ShowStock:      false,
			Status:         FrontProductVisibilityStatusHidden,
			ReasonCode:     FrontProductVisibilityReasonRejected,
			ReasonMessage:  "商品审核未通过",
			RecoveryHint:   "请浏览其他在售商品，或联系运营确认商品状态",
			FallbackAction: FrontProductFallbackBrowseSimilar,
			FallbackTarget: "product_list",
		}
	default:
		return ProductVisibilityResult{
			Visible:        false,
			Purchasable:    false,
			ShowPrice:      false,
			ShowStock:      false,
			Status:         FrontProductVisibilityStatusHidden,
			ReasonCode:     FrontProductVisibilityReasonPendingReview,
			ReasonMessage:  "商品暂未开放浏览",
			RecoveryHint:   "请稍后再试，或浏览当前分类下的其他商品",
			FallbackAction: FrontProductFallbackBrowseProductList,
			FallbackTarget: "product_list",
		}
	}

	if product.PublishStatus != 1 {
		return ProductVisibilityResult{
			Visible:        false,
			Purchasable:    false,
			ShowPrice:      false,
			ShowStock:      false,
			Status:         FrontProductVisibilityStatusHidden,
			ReasonCode:     FrontProductVisibilityReasonOffShelf,
			ReasonMessage:  "商品已下架",
			RecoveryHint:   "请返回商品列表挑选其他仍在售的商品",
			FallbackAction: FrontProductFallbackBrowseProductList,
			FallbackTarget: "product_list",
		}
	}

	if product.PreviewStatus != 0 {
		return ProductVisibilityResult{
			Visible:        false,
			Purchasable:    false,
			ShowPrice:      false,
			ShowStock:      false,
			Status:         FrontProductVisibilityStatusHidden,
			ReasonCode:     FrontProductVisibilityReasonPreviewOnly,
			ReasonMessage:  "商品暂未开放浏览",
			RecoveryHint:   "当前商品仍处于预告阶段，请稍后关注正式上架状态",
			FallbackAction: FrontProductFallbackBrowseProductList,
			FallbackTarget: "product_list",
		}
	}

	if product.Stock <= 0 {
		return ProductVisibilityResult{
			Visible:        true,
			Purchasable:    false,
			ShowPrice:      true,
			ShowStock:      true,
			Status:         FrontProductVisibilityStatusNotSaleable,
			ReasonCode:     FrontProductVisibilityReasonSoldOut,
			ReasonMessage:  "商品暂时缺货",
			RecoveryHint:   "可先浏览同类商品，或等待商品补货后再购买",
			FallbackAction: FrontProductFallbackBrowseSimilar,
			FallbackTarget: "product_list",
		}
	}

	return ProductVisibilityResult{
		Visible:        true,
		Purchasable:    true,
		ShowPrice:      true,
		ShowStock:      true,
		Status:         FrontProductVisibilityStatusVisible,
		ReasonCode:     FrontProductVisibilityReasonNone,
		ReasonMessage:  "",
		RecoveryHint:   "",
		FallbackAction: "",
		FallbackTarget: "",
	}
}

func EnsureFrontProductVisible(product *pmsclient.ProductSpuListData) error {
	result := BuildFrontProductVisibility(product)
	if result.Visible {
		return nil
	}

	return errors.New(result.ReasonMessage)
}

func BuildOwnerDisabledVisibility() ProductVisibilityResult {
	return ProductVisibilityResult{
		Visible:        false,
		Purchasable:    false,
		ShowPrice:      false,
		ShowStock:      false,
		Status:         FrontProductVisibilityStatusHidden,
		ReasonCode:     FrontProductVisibilityReasonOwnerDisabled,
		ReasonMessage:  "商品所属主体已停用",
		RecoveryHint:   "请返回商品列表浏览其他仍可售的商品",
		FallbackAction: FrontProductFallbackBrowseProductList,
		FallbackTarget: "product_list",
	}
}
