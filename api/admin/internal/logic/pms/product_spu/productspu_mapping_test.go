package product_spu

import (
	"testing"

	"github.com/feihua/zero-admin/api/admin/internal/types"
)

func TestBuildProductSpuNestedListsPreserveDetailIDs(t *testing.T) {
	req := &types.AddProductSpuReq{
		MemberPriceList: []types.MemberPriceReq{{
			Id:              11,
			MemberLevelId:   101,
			MemberPrice:     8800,
			MemberLevelName: "黄金会员",
		}},
		AttributeValueList: []types.AddProductAttributeValueReq{{
			Id:          12,
			AttributeId: 202,
			Value:       "黑色",
			Status:      1,
		}},
		FullList: []types.ProductFullReductionReq{{
			Id:          13,
			FullPrice:   10000,
			ReducePrice: 1000,
		}},
		LadderList: []types.ProductLadderReq{{
			Id:       14,
			Count:    3,
			Discount: 80,
			Price:    7600,
		}},
		SkuList: []types.AddProductSkuReq{{
			Id:       15,
			SpuId:    2001,
			Name:     "黑-L",
			SkuCode:  "SKU-001",
			MainPic:  "main.png",
			AlbumPics:"a.png,b.png",
			Price:    99.5,
			Stock:    10,
			LowStock: 2,
			SpecData: `{"颜色":"黑","尺码":"L"}`,
			Weight:   1.2,
		}},
	}

	if got := buildMemberPriceList(req); len(got) != 1 || got[0].Id != 11 {
		t.Fatalf("member price id not preserved: %+v", got)
	}
	if got := buildProductAttributeValueList(req); len(got) != 1 || got[0].Id != 12 {
		t.Fatalf("attribute value id not preserved: %+v", got)
	}
	if got := buildProductFullReductionList(req); len(got) != 1 || got[0].Id != 13 {
		t.Fatalf("full reduction id not preserved: %+v", got)
	}
	if got := buildProductLadderList(req); len(got) != 1 || got[0].Id != 14 {
		t.Fatalf("ladder id not preserved: %+v", got)
	}
	if got := buildSkuStockList(req); len(got) != 1 || got[0].Id != 15 || got[0].SpuId != 2001 {
		t.Fatalf("sku detail mapping not preserved: %+v", got)
	}
}

func TestBuildUpdateProductSpuNestedListsPreserveDetailIDs(t *testing.T) {
	req := &types.UpdateProductSpuReq{
		MemberPriceList: []types.MemberPriceReq{{
			Id:              21,
			MemberLevelId:   301,
			MemberPrice:     6600,
			MemberLevelName: "白银会员",
		}},
		AttributeValueList: []types.AddProductAttributeValueReq{{
			Id:          22,
			AttributeId: 402,
			Value:       "白色",
			Status:      1,
		}},
		FullList: []types.ProductFullReductionReq{{
			Id:          23,
			FullPrice:   20000,
			ReducePrice: 1500,
		}},
		LadderList: []types.ProductLadderReq{{
			Id:       24,
			Count:    2,
			Discount: 90,
			Price:    8800,
		}},
		SkuList: []types.AddProductSkuReq{{
			Id:       25,
			SpuId:    3001,
			Name:     "白-M",
			SkuCode:  "SKU-002",
			MainPic:  "main-2.png",
			AlbumPics:"c.png,d.png",
			Price:    129.5,
			Stock:    6,
			LowStock: 1,
			SpecData: `{"颜色":"白","尺码":"M"}`,
			Weight:   1.1,
		}},
	}

	if got := buildUpdateMemberPriceList(req); len(got) != 1 || got[0].Id != 21 {
		t.Fatalf("update member price id not preserved: %+v", got)
	}
	if got := buildUpdateProductAttributeValueList(req); len(got) != 1 || got[0].Id != 22 {
		t.Fatalf("update attribute value id not preserved: %+v", got)
	}
	if got := buildUpdateProductFullReductionList(req); len(got) != 1 || got[0].Id != 23 {
		t.Fatalf("update full reduction id not preserved: %+v", got)
	}
	if got := buildUpdateProductLadderList(req); len(got) != 1 || got[0].Id != 24 {
		t.Fatalf("update ladder id not preserved: %+v", got)
	}
	if got := buildUpdateSkuStockList(req); len(got) != 1 || got[0].Id != 25 || got[0].SpuId != 3001 {
		t.Fatalf("update sku detail mapping not preserved: %+v", got)
	}
}
