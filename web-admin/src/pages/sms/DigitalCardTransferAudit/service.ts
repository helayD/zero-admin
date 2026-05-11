import { request } from 'umi';
import type {
  QueryDigitalCardTransferLogListParams,
  QueryDigitalCardTransferLogListResp,
} from './data.d';

// Story 10.7 Task 9.2 — 后台合规审计跨资产检索
export async function queryDigitalCardTransferLogList(
  params: QueryDigitalCardTransferLogListParams,
) {
  return request<QueryDigitalCardTransferLogListResp>(
    '/api/sms/digitalCardAsset/queryDigitalCardTransferLogList',
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
