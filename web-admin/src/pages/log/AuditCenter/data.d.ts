export interface AuditCenterItem {
  sourceType: string;
  sourceId: number;
  traceId: string;
  eventType: string;
  action: string;
  result: string;
  scopeType: string;
  scopeLabel: string;
  platformId: number;
  tenantId: number;
  merchantId: number;
  operatorId: number;
  operatorName: string;
  resourceType: string;
  resourceId: number;
  resourceName: string;
  subjectInfo: string;
  requestSummary: string;
  sensitiveMasked: boolean;
  happenedAt: string;
}

export interface AuditCenterListParams {
  current?: number;
  pageSize?: number;
  scopeType?: string;
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
  sourceType?: string;
  traceId?: string;
  operatorId?: number;
  operatorName?: string;
  resourceType?: string;
  resourceId?: number;
  eventType?: string;
  result?: string;
  keyword?: string;
  startTime?: string;
  endTime?: string;
}

export interface AuditTimelineItem extends AuditCenterItem {}

export interface AuditCenterDetailData extends AuditCenterItem {
  detailPayload: string;
  timeline: AuditTimelineItem[];
}
