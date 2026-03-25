export interface PreferredAreaListItem {
  id?: number;
  name?: string;
  subTitle?: string;
  pic?: string;
  sort?: number;
  showStatus?: number;
  createBy?: string;
  createTime?: string;
  updateBy?: string;
  updateTime?: string;
  effectiveStatus?: string;
}

export interface PreferredAreaListPagination {
  total: number;
  pageSize: number;
  current: number;
}

export interface PreferredAreaListData {
  list: PreferredAreaListItem[];
  pagination: Partial<PreferredAreaListPagination>;
}

export interface PreferredAreaListParams {
  name?: string;
  showStatus?: number;
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
  pageSize?: number;
  current?: number;
  filter?: { [key: string]: any[] };
  sorter?: { [key: string]: any };
}
