export interface HomeRecommendProductListItem {
  id: number;
  productId: number;
  recommendStatus: number;
  productName: string;
}

export interface ProductListItem {
  id: number;

}

export interface HomeRecommendProductListPagination {
  total: number;
  pageSize: number;
  current: number;
}

export interface HomeRecommendProductListData {
  list: HomeRecommendProductListItem[];
  pagination: Partial<HomeRecommendProductListPagination>;
}

export interface HomeRecommendProductListParams {
  recommendStatus?: number;
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
  pageSize?: number;
  currentPage?: number;
  filter?: { [key: string]: any[] };
  sorter?: { [key: string]: any };
}
