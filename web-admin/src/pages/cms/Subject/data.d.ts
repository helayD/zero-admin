export interface SubjectListItem {
  id: number;
  categoryId?: number;
  title?: string;
  pic?: string;
  productCount?: number;
  recommendStatus?: number;
  collectCount?: number;
  readCount?: number;
  commentCount?: number;
  albumPics?: string;
  description?: string;
  showStatus?: number;
  content?: string;
  forwardCount?: number;
  categoryName?: string;
  createBy?: string;
  createTime?: string;
  updateBy?: string;
  updateTime?: string;
  sort?: number;
}

export interface SubjectListPagination {
  total: number;
  pageSize: number;
  current: number;
}

export interface SubjectListData {
  list: SubjectListItem[];
  pagination: Partial<SubjectListPagination>;
}

export interface SubjectListParams {
  title?: string;
  recommendStatus?: number;
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
