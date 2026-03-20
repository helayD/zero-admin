export interface PostListItem {
  id: number;
  postCode?: string;
  postName?: string;
  sort?: number;
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

export interface PostListPagination {
  total: number;
  pageSize: number;
  current: number;
}

export interface PostListData {
  list: PostListItem[];
  pagination: Partial<PostListPagination>;
}

export interface PostListParams {

  delFlag?: number;
  scopeType?: string;
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
  pageSize?: number;
  current?: number;
  filter?: { [key: string]: any[] };
  sorter?: { [key: string]: any };

}
