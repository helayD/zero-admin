export interface ProductSpuListItem {
  id: number; //商品SpuId
  name: string; //商品名称
  productSn: string; //商品货号
  categoryId: number; //商品分类ID
  categoryIds: string; //商品分类ID集合
  categoryName: string; //商品分类名称
  brandId: number; //品牌ID
  brandName: string; //品牌名称
  unit: string; //单位
  weight: number; //重量(kg)
  keywords: string; //关键词
  brief: string; //简介
  description: string; //详细描述
  albumPics: string; //画册图片，最多8张，以逗号分割
  mainPic: string; //主图
  priceRange: string; //价格区间
  publishStatus: number; //上架状态：0-下架，1-上架
  newStatus: number; //新品状态:0->不是新品；1->新品
  recommendStatus: number; //推荐状态；0->不推荐；1->推荐
  verifyStatus: number; //审核状态：0->未审核；1->审核通过
  previewStatus: number; //是否为预告商品：0->不是；1->是
  sort: number; //排序
  newStatusSort: number; //新品排序
  recommendStatusSort: number; //推荐排序
  sales: number; //销量
  stock: number; //库存
  lowStock: number; //预警库存
  promotionType: number; //促销类型：0->没有促销使用原价;1->使用促销价；2->使用会员价；3->使用阶梯价格；4->使用满减价格；5->秒杀
  subTitle: string; //副标题
  detailTitle: string; //详情标题
  detailDesc: string; //详情描述
  detailHtml: string; //产品详情网页内容
  detailMobileHtml: string; //移动端网页详情
  createBy: number; //创建人ID
  createTime: string; //创建时间
  updateBy: number; //更新人ID
  updateTime: string; //更新时间
  scopeType?: 'platform' | 'tenant' | 'merchant'; //作用域来源
  platformId?: number; //平台ID
  tenantId?: number; //租户ID
  merchantId?: number; //商户ID
  reviewMan?: string; //最近审核人
  reviewTime?: string; //最近审核时间
  reviewDetail?: string; //最近审核意见
  publishMan?: string; //最近上下架操作人
  publishTime?: string; //最近上下架时间
  publishDetail?: string; //最近上下架说明
  recommendMan?: string; //最近推荐操作人
  recommendTime?: string; //最近推荐时间
  recommendDetail?: string; //最近推荐说明
  isDeleted: number; //是否删除
  fulfillmentMode?: 'physical_delivery' | 'digital_asset'; //履约模式：physical_delivery-实物发货，digital_asset-数字资产
  fulfillmentRuleId?: number; //发卡规则ID（仅digital_asset模式有效）
}

export interface ProductSpuListPagination {
  total: number;
  pageSize: number;
  current: number;
}

export interface ProductSpuSkuItem {
  id?: number;
  spuId?: number;
  name?: string;
  skuCode?: string;
  mainPic?: string;
  albumPics?: string;
  price?: number;
  promotionPrice?: number;
  promotionStartTime?: string;
  promotionEndTime?: string;
  stock?: number;
  lowStock?: number;
  specData?: string;
  weight?: number;
  publishStatus?: number;
  verifyStatus?: number;
  sort?: number;
}

export interface ProductSpuMemberPriceItem {
  id?: number;
  memberLevelId?: number;
  memberLevelName?: string;
  memberPrice?: number;
}

export interface ProductSpuLadderItem {
  id?: number;
  count?: number;
  discount?: number;
  price?: number;
}

export interface ProductSpuFullReductionItem {
  id?: number;
  fullPrice?: number;
  reducePrice?: number;
}

export interface ProductSpuAttributeValueItem {
  id?: number;
  spuId?: number;
  attributeId?: number;
  value?: string;
  status?: number;
}

export interface ProductSpuNestedPayload {
  ladderList?: ProductSpuLadderItem[];
  fullList?: ProductSpuFullReductionItem[];
  memberPriceList?: ProductSpuMemberPriceItem[];
  skuList?: ProductSpuSkuItem[];
  attributeValueList?: ProductSpuAttributeValueItem[];
  subjectIds?: number[];
  prefrenceAreaIds?: number[];
}

export type ProductSpuSubmitPayload = Partial<ProductSpuListItem> &
  ProductSpuNestedPayload & {
    scopeType?: 'platform' | 'tenant' | 'merchant';
    platformId?: number;
    tenantId?: number;
    merchantId?: number;
  };

export type ProductSpuDraftFormValues = ProductSpuSubmitPayload;

export interface ProductSpuDetailPayload {
  productData?: ProductSpuDraftFormValues & Record<string, any>;
  ladderList?: ProductSpuLadderItem[];
  fullList?: ProductSpuFullReductionItem[];
  memberPriceList?: ProductSpuMemberPriceItem[];
  skuList?: ProductSpuSkuItem[];
  attributeValueList?: ProductSpuAttributeValueItem[];
  subjectIds?: number[];
  prefrenceAreaIds?: number[];
  couponList?: Record<string, any>[];
}

export interface ProductSpuDetailResponse {
  code: string;
  message: string;
  data: ProductSpuDetailPayload;
}

export interface ProductSpuListData {
  list: ProductSpuListItem[];
  pagination: Partial<ProductSpuListPagination>;
}

export interface ProductSpuListParams {
  id?: number; //商品SpuId
  name?: string; //商品名称
  productSn?: string; //商品货号
  categoryId?: number; //商品分类ID
  categoryIds?: string; //商品分类ID集合
  categoryName?: string; //商品分类名称
  brandId?: number; //品牌ID
  brandName?: string; //品牌名称
  unit?: string; //单位
  weight?: number; //重量(kg)
  keywords?: string; //关键词
  brief?: string; //简介
  description?: string; //详细描述
  albumPics?: string; //画册图片，最多8张，以逗号分割
  mainPic?: string; //主图
  priceRange?: string; //价格区间
  publishStatus?: number; //上架状态：0-下架，1-上架
  newStatus?: number; //新品状态:0->不是新品；1->新品
  recommendStatus?: number; //推荐状态；0->不推荐；1->推荐
  verifyStatus?: number; //审核状态：0->未审核；1->审核通过
  previewStatus?: number; //是否为预告商品：0->不是；1->是
  sort?: number; //排序
  newStatusSort?: number; //新品排序
  recommendStatusSort?: number; //推荐排序
  sales?: number; //销量
  stock?: number; //库存
  lowStock?: number; //预警库存
  promotionType?: number; //促销类型：0->没有促销使用原价;1->使用促销价；2->使用会员价；3->使用阶梯价格；4->使用满减价格；5->秒杀
  subTitle?: string; //副标题
  detailTitle?: string; //详情标题
  detailDesc?: string; //详情描述
  detailHtml?: string; //产品详情网页内容
  detailMobileHtml?: string; //移动端网页详情
  createBy?: number; //创建人ID
  createTime?: string; //创建时间
  updateBy?: number; //更新人ID
  updateTime?: string; //更新时间
  reviewMan?: string;
  reviewTime?: string;
  reviewDetail?: string;
  publishMan?: string;
  publishTime?: string;
  publishDetail?: string;
  recommendMan?: string;
  recommendTime?: string;
  recommendDetail?: string;
  isDeleted?: number; //是否删除
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;

  pageSize?: number;
  current?: number;
  filter?: Record<string, any[]>;
  sorter?: Record<string, any>;
}
