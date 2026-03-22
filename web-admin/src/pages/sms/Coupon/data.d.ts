export interface CouponListItem {
  id?: number;
  name?: string;
  type?: number;
  typeId?: number;
  platform?: number;
  useType?: number;
  status?: number;
  isEnabled?: number;
  enableTime?: any;
  startTime?: any;
  endTime?: string;
  amount?: number;
  minPoint?: number;
  perLimit?: number;
  note?: string;
  publishCount?: number;
  receiveCount?: number;
  useCount?: number;
  memberLevel?: number;
  code?: string;
  productCategoryRelationList?: any[];
  productRelationList?: any[];
}

export interface CouponListPagination {
  total: number;
  pageSize: number;
  current: number;
}

export interface CouponListData {
  list: CouponListItem[];
  pagination: Partial<CouponListPagination>;
}

export interface CouponListParams {
  type?: number;
  platform?: number;
  useType?: number;
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
  /** 当前的页码 */
  current?: number;
  /** 页面的容量 */
  pageSize?: number;
}

export interface ProductListItem {
  id: number;
  name?: string;
  productSn?: string;
  promotionPrice?: number | string;
  originalPrice?: number | string;
  stock?: number;
}

export interface CategoryListItem {
  id: number;
  name?: string;
  icon?: string;
  description?: string;
  parentId?: number;
}

export interface CategoryListParams {


  parentId?: number;
  pageSize?: number;
  current?: number;
  filter?: { [key: string]: any[] };
  sorter?: { [key: string]: any };

}
export interface CouponHistoryListItem {
  id: number;
  useStatus: number;
  getType: number;
  couponCode?: string;
  memberNickName?: string;
  useTime?: string;
  orderId?: number;
}

export interface CouponHistoryListParams {
  id?: number;
  useStatus?: number;
  couponId?: number;
  useType?: number;
  current?: number;
  pageSize?: number;
}

export interface CouponDetailData {
  id: number;
  type: number;
  name: string;
  platform: number;
  count: number;
  amount: number;
  perLimit: number;
  minPoint: number;
  startTime: string;
  endTime: string;
  useType: number;
  note: string;
  publishCount: number;
  useCount: number;
  receiveCount: number;
  enableTime: string;
  code: string;
  memberLevel: number;
}
