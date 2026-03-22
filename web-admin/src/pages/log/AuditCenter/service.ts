import {request} from 'umi';
import type { AuditCenterListParams } from './data.d';

export async function queryAuditCenterList(params: AuditCenterListParams) {
  return request('/api/sys/log/queryAuditCenterList', { method: 'GET', params });
}

export async function queryAuditCenterDetail(params: {
  sourceType: string;
  sourceId: number;
  scopeType?: string;
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
}) {
  return request('/api/sys/log/queryAuditCenterDetail', { method: 'GET', params });
}
