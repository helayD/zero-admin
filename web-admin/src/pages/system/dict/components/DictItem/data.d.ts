export interface DictItemListItem {
  id: number;
  status?: number;
  dictType?: string;
  dictName?: string;
  isDefault?: string;
  dictTypeId?: number;
  dictLabel?: string;
  dictValue?: string;
  dictSort?: number;
  cssClass?: string;
  listClass?: string;
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

export interface DictItemListPagination {
  total: number;
  pageSize: number;
  current: number;
}

export interface DictItemListData {
  list: DictItemListItem[];
  pagination: Partial<DictItemListPagination>;
}

export interface DictItemListParams {
  dictType?:string;
  dictTypeId?: number;
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
