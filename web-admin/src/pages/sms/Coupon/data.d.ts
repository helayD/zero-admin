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
  minAmount?: number;
  minPoint?: number;
  perLimit?: number;
  note?: string;
  description?: string;
  totalCount?: number;
  publishCount?: number;
  receivedCount?: number;
  receiveCount?: number;
  usedCount?: number;
  useCount?: number;
  memberLevel?: number;
  code?: string;
  scopeCount?: number;
  typeName?: string;
  createBy?: number;
  createTime?: string;
  updateBy?: number;
  updateTime?: string;
  scopeType?: number | string;
  couponScopeData?: any[];
  productCategoryRelationList?: any[];
  productRelationList?: any[];
  effectiveStatus?: string;
  scopeSummary?: string;
  affectedPaths?: string;
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
  typeId: number;
  typeName?: string;
  name: string;
  code: string;
  amount: number;
  minAmount: number;
  startTime: string;
  endTime: string;
  totalCount: number;
  receivedCount: number;
  usedCount: number;
  perLimit: number;
  status: number;
  isEnabled: number;
  description: string;
  createBy: number;
  createTime: string;
  updateBy: number;
  updateTime: string;
  scopeType: number;
  couponScopeData?: any[];
  scopeCount?: number;
  effectiveStatus?: string;
  scopeSummary?: string;
  affectedPaths?: string;
  // 兼容老字段
  type?: number;
  platform?: number;
  useType?: number;
  minPoint?: number;
  note?: string;
  publishCount?: number;
  useCount?: number;
  receiveCount?: number;
  enableTime?: string;
  memberLevel?: number;
  count?: number;
}
