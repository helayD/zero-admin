export interface UserListItem {
  id: number;
  name?: string;
  userName?: string;
  nickName?: string;
  deptId?: number | string;
  deptName?: string;
  postIds?: number[];
  roleIds?: number[];
  status?: number;
  mobile?: string;
  email?: string;
  password?: string;
  userType?: string;
  remark?: string;
  createBy?: string;
  createTime?: string;
  updateBy?: string;
  updateTime?: string;
  loginIp?: string;
  loginDate?: string;
  loginBrowser?: string;
  loginOs?: string;
  pwdUpdateDate?: string;
  scopeType?: string;
  scopeLabel?: string;
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
  activationStatus?: string;
  roleMode?: string;
}

export interface UserListPagination {
  total: number;
  pageSize: number;
  current: number;
}

export interface UserListData {
  list: UserListItem[];
  pagination: Partial<UserListPagination>;
}

export interface UserListParams {
  id?: number;
  deptId?: number;
  status?: number;
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

export interface RoleList {
  id: number;
  name: string;
  remark: string;
}


export interface JobList {
  id: number;
  postName: string;
}

export interface SelectData {
  roleList: RoleList[];
  jobList: JobList[];
}
