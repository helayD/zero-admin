import { request } from 'umi';
import type {
  MerchantOrderItem,
  MerchantOrderListParams,
  CompanyAddressItem,
  CompanyAddressListParams,
  OrderOperationLogItem,
  OrderOperationLogListParams,
  ConfirmDeliveryReq,
  UpdateOrderReturnReq,
} from './data.d';

export async function queryMerchantOrderList(params: MerchantOrderListParams) {
  return request<{
    code: string;
    message: string;
    total: number;
    current: number;
    pageSize: number;
    data: MerchantOrderItem[];
    success: boolean;
  }>('/api/oms/merchantOrder/list', {
    method: 'POST',
    data: params,
  });
}

export async function confirmMerchantDelivery(params: ConfirmDeliveryReq) {
  return request<{ code: string; message: string }>('/api/oms/merchantDelivery/confirm', {
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

export async function updateMerchantOrderRemark(params: { id: number; note: string }) {
  return request<{ code: string; message: string }>('/api/oms/order/updateNote', {
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
