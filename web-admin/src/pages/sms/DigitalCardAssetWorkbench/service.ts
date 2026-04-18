import { request } from 'umi';
import type {
  DigitalCardAssetActionResp,
  DigitalCardAssetListParams,
  QueryDigitalCardAssetDetailResp,
  QueryDigitalCardAssetListResp,
} from './data';
import { normalizeDigitalCardAssetListResponse } from './helper';

type GovernancePayload = {
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
};

type ActionPayload = GovernancePayload & {
  assetInstanceId: number;
  reason: string;
};

export async function queryDigitalCardAssetList(params: DigitalCardAssetListParams) {
  const response = await request<QueryDigitalCardAssetListResp>(
    '/api/sms/digitalCardAsset/queryDigitalCardAssetList',
    {
      method: 'GET',
      params: {
        ...params,
        current: params.current ?? 1,
        pageSize: params.pageSize ?? 20,
      },
    },
  );

  return normalizeDigitalCardAssetListResponse(response);
}

export async function queryDigitalCardAssetDetail(assetInstanceId: number) {
  return request<QueryDigitalCardAssetDetailResp>(
    '/api/sms/digitalCardAsset/queryDigitalCardAssetDetail',
    {
      method: 'GET',
      params: { assetInstanceId },
    },
  );
}

export async function reviewDigitalCardAssetCompliance(payload: ActionPayload) {
  return request<DigitalCardAssetActionResp>(
    '/api/sms/digitalCardAsset/reviewDigitalCardAssetCompliance',
    {
      method: 'POST',
      data: payload,
    },
  );
}

export async function offlineDigitalCardAssetDisplay(payload: ActionPayload) {
  return request<DigitalCardAssetActionResp>(
    '/api/sms/digitalCardAsset/offlineDigitalCardAssetDisplay',
    {
      method: 'POST',
      data: payload,
    },
  );
}

export async function recycleDigitalCardAsset(payload: ActionPayload) {
  return request<DigitalCardAssetActionResp>(
    '/api/sms/digitalCardAsset/recycleDigitalCardAsset',
    {
      method: 'POST',
      data: payload,
    },
  );
}
