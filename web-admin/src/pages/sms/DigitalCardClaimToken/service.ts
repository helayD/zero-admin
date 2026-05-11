import { request } from 'umi';
import type {
  AdminRevokeDigitalCardClaimTokenParams,
  AdminRevokeDigitalCardClaimTokenResp,
  QueryDigitalCardClaimTokenListParams,
  QueryDigitalCardClaimTokenListResp,
} from './data.d';

// Story 10.7 Task 8.8 / S5+S6 — 后台分享凭证管理
export async function queryDigitalCardClaimTokenList(
  params: QueryDigitalCardClaimTokenListParams,
) {
  return request<QueryDigitalCardClaimTokenListResp>(
    '/api/sms/digitalCardAsset/queryDigitalCardClaimTokenList',
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

// Story 10.7 Task 8.8 / S5+S6 — 后台手动吊销分享凭证
export async function adminRevokeDigitalCardClaimToken(
  data: AdminRevokeDigitalCardClaimTokenParams,
) {
  return request<AdminRevokeDigitalCardClaimTokenResp>(
    '/api/sms/digitalCardAsset/adminRevokeDigitalCardClaimToken',
    {
      method: 'POST',
      data,
    },
  );
}
