export interface DictTypeListItem {
  id: number;
  status?: number;
  dictType?: string;
  dictName?: string;
  remark?: string;
  createBy?: string;
  createTime?: string;
  updateBy?: string;
  updateTime?: string;
  scopeType?: string;
  scopeLabel?: string;
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
}

export interface DictTypeListPagination {
  total: number;
  pageSize: number;
  current: number;
}

export interface DictTypeListData {
  list: DictTypeListItem[];
  pagination: Partial<DictTypeListPagination>;
}

export interface DictTypeListParams {
  dictName?: string;
  dictType?: string;
  status?: number;
  scopeType?: string;
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
  pageSize?: number;
  current?: number;
  filter?: { [key: string]: any[] };
  sorter?: { [key: string]: any };

}
