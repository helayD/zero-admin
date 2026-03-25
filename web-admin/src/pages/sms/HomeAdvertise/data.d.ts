export interface HomeAdvertiseListItem {
  id: number;
  status: number;
  name: string;
  startTime: string;
  endTime: string;
  pic: string;
  type: number;
  url: string;
  remark: string;
  sort: number;
  clickCount: number;
  orderCount: number;
  createTime: string;
  updateTime: string;
}

export interface HomeAdvertiseListPagination {
  total: number;
  pageSize: number;
  current: number;
}

export interface HomeAdvertiseListData {
  list: HomeAdvertiseListItem[];
  pagination: Partial<HomeAdvertiseListPagination>;
}

export interface HomeAdvertiseListParams {
  type?: number;
  status?: number;
  pageSize?: number;
  currentPage?: number;
  filter?: { [key: string]: any[] };
  sorter?: { [key: string]: any };

}
