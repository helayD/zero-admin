export interface DeptListItem {
  id: number;
  parentId?: number;
  deptName?: string;
  sort?: number;
  status?: number;
  email?: string;
  leader?: string;
  phone?: string;
  remark?: string;
  createBy?: string;
  createTime?: string;
  updateBy?: string;
  updateTime?: string;
  ancestors?: string;
  scopeType?: string;
  scopeLabel?: string;
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
}

export interface DeptListPagination {
  total: number;
  pageSize: number;
  current: number;
}

export interface DeptListData {
  list: DeptListItem[];
  pagination: Partial<DeptListPagination>;
}

export interface DeptListParams {

  scopeType?: string;
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
  pageSize?: number;
  current?: number;
  currentPage?: number;
  filter?: { [key: string]: any[] };
  sorter?: { [key: string]: any };

}
