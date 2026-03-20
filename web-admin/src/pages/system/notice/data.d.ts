export interface NoticeListItem {
  id: number;
  noticeTitle?: string;
  noticeType?: number;
  noticeContent?: string;
  status?: number;
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

export interface NoticeListPagination {
  total: number;
  pageSize: number;
  current: number;
}

export interface NoticeListData {
  list: NoticeListItem[];
  pagination: Partial<NoticeListPagination>;
}

export interface NoticeListParams {
  noticeTitle?: string;
  noticeType?: number;
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
