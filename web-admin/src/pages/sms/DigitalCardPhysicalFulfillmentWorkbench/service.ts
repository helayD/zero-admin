import { request } from 'umi';
import type {
  DigitalCardPhysicalFulfillmentActionResp,
  DigitalCardPhysicalFulfillmentListParams,
  QueryDigitalCardPhysicalFulfillmentDetailResp,
  QueryDigitalCardPhysicalFulfillmentListResp,
} from './data';
import { normalizePhysicalFulfillmentListResponse } from './helper';

type GovernancePayload = {
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
};

export type EnsurePayload = GovernancePayload & {
  assetInstanceId: number;
};

export type ProductionPayload = GovernancePayload & {
  fulfillmentId: number;
  productionBatchNo?: string;
  productionStatus: string;
  reason: string;
};

export type ShipPayload = GovernancePayload & {
  fulfillmentId: number;
  carrierCode: string;
  carrierName: string;
  trackingNo: string;
  reason: string;
};

export type ExceptionPayload = GovernancePayload & {
  fulfillmentId: number;
  failureCode?: string;
  reason: string;
};

export async function queryDigitalCardPhysicalFulfillmentList(
  params: DigitalCardPhysicalFulfillmentListParams,
) {
  const response = await request<QueryDigitalCardPhysicalFulfillmentListResp>(
    '/api/sms/digitalCardPhysicalFulfillment/queryDigitalCardPhysicalFulfillmentList',
    {
      method: 'GET',
      params: {
        ...params,
        current: params.current ?? 1,
        pageSize: params.pageSize ?? 20,
      },
    },
  );

  return normalizePhysicalFulfillmentListResponse(response);
}

export async function queryDigitalCardPhysicalFulfillmentDetail(fulfillmentId: number) {
  return request<QueryDigitalCardPhysicalFulfillmentDetailResp>(
    '/api/sms/digitalCardPhysicalFulfillment/queryDigitalCardPhysicalFulfillmentDetail',
    {
      method: 'GET',
      params: { fulfillmentId },
    },
  );
}

export async function ensureDigitalCardPhysicalFulfillment(payload: EnsurePayload) {
  return request<DigitalCardPhysicalFulfillmentActionResp>(
    '/api/sms/digitalCardPhysicalFulfillment/ensureDigitalCardPhysicalFulfillment',
    {
      method: 'POST',
      data: payload,
    },
  );
}

export async function updateDigitalCardPhysicalProductionStatus(payload: ProductionPayload) {
  return request<DigitalCardPhysicalFulfillmentActionResp>(
    '/api/sms/digitalCardPhysicalFulfillment/updateDigitalCardPhysicalProductionStatus',
    {
      method: 'POST',
      data: payload,
    },
  );
}

export async function shipDigitalCardPhysicalFulfillment(payload: ShipPayload) {
  return request<DigitalCardPhysicalFulfillmentActionResp>(
    '/api/sms/digitalCardPhysicalFulfillment/shipDigitalCardPhysicalFulfillment',
    {
      method: 'POST',
      data: payload,
    },
  );
}

export async function markDigitalCardPhysicalFulfillmentException(payload: ExceptionPayload) {
  return request<DigitalCardPhysicalFulfillmentActionResp>(
    '/api/sms/digitalCardPhysicalFulfillment/markDigitalCardPhysicalFulfillmentException',
    {
      method: 'POST',
      data: payload,
    },
  );
}

export async function requestDigitalCardPhysicalReissue(payload: ExceptionPayload) {
  return request<DigitalCardPhysicalFulfillmentActionResp>(
    '/api/sms/digitalCardPhysicalFulfillment/requestDigitalCardPhysicalReissue',
    {
      method: 'POST',
      data: payload,
    },
  );
}
