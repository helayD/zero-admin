export interface SubjectCategoryListItem {
  id?: number;
  name?: string;
  icon?: string;
  subjectCount?: number;
  showStatus?: number;
  sort?: number;
  createBy?: string;
  createTime?: string;
  updateBy?: string;
  updateTime?: string;
}

export interface SubjectCategoryListPagination {
  total: number;
  pageSize: number;
  current: number;
}

export interface SubjectCategoryListData {
  list: SubjectCategoryListItem[];
  pagination: Partial<SubjectCategoryListPagination>;
}

export interface SubjectCategoryListParams {
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
