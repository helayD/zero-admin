import { request } from 'umi';
import type {
  QueryDigitalCardRedemptionOrderListParams,
  QueryDigitalCardRedemptionOrderListResp,
} from './data.d';

// Story 10.7 Task 2.8 / S3+S4 — 后台提货单管理
export async function queryDigitalCardRedemptionOrderList(
  params: QueryDigitalCardRedemptionOrderListParams,
) {
  return request<QueryDigitalCardRedemptionOrderListResp>(
    '/api/sms/digitalCardAsset/queryDigitalCardRedemptionOrderList',
    {
      method: 'GET',
      params: {
        ...params,
        current: params.current ?? 1,
        pageSize: params.pageSize ?? 20,
      },
    },
  );
}
