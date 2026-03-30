import { request } from 'umi';
import type {
  CustomerServiceOrderItem,
  CustomerServiceOrderListParams,
  CompanyAddressItem,
  CompanyAddressListParams,
  OrderOperationLogItem,
  OrderOperationLogListParams,
  UpdateOrderReturnReq,
} from './data.d';

type GovernancePayload = {
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
};

export async function queryCustomerServiceOrderList(params: CustomerServiceOrderListParams) {
  return request<{
    code: string;
    message: string;
    total: number;
    current: number;
    pageSize: number;
    data: CustomerServiceOrderItem[];
    success: boolean;
  }>('/api/oms/customerServiceOrderList', {
    method: 'POST',
    data: params,
  });
}

export async function queryCompanyAddressList(params: CompanyAddressListParams) {
  return request<{
    code: string;
    message: string;
    total: number;
    current: number;
    pageSize: number;
    data: CompanyAddressItem[];
    success: boolean;
  }>('/api/oms/companyAddress/list', {
    method: 'POST',
    data: params,
  });
}

export async function queryOrderOperationLogList(params: OrderOperationLogListParams) {
  return request<{
    code: string;
    message: string;
    total: number;
    current: number;
    pageSize: number;
    data: OrderOperationLogItem[];
    success: boolean;
  }>('/api/oms/orderOperationLog/list', {
    method: 'POST',
    data: params,
  });
}

export async function updateOrderReturn(params: UpdateOrderReturnReq) {
  return request<{ code: string; message: string }>('/api/oms/orderReturn/updateOrderReturn', {
    method: 'POST',
    data: params,
  });
}
