export interface TenantListItem {
  id: number;
  tenantCode: string;
  tenantName: string;
  tenantShortName?: string;
  contactName: string;
  contactMobile: string;
  contactEmail?: string;
  availableChannels: string[];
  dataRetentionDays: number;
  featureFlags: string[];
  status: number;
  statusReason?: string;
  primaryAdminUserId: number;
  primaryAdminUserName?: string;
  primaryAdminMobile?: string;
  adminActivationStatus: string;
  createdBy?: string;
  createdAt?: string;
  updatedBy?: string;
  updatedAt?: string;
}

export interface TenantCreateParams {
  tenantName: string;
  tenantShortName?: string;
  contactName: string;
  contactMobile: string;
  contactEmail?: string;
  availableChannels: string[];
  dataRetentionDays: number;
  featureFlags: string[];
  adminUserName: string;
  adminNickName: string;
  adminMobile: string;
  adminEmail?: string;
  adminPassword: string;
}

export interface CreateTenantResult {
  tenantId: number;
  tenantCode: string;
  adminUserId: number;
  adminActivationStatus: string;
}

export interface TenantListParams {
  current?: number;
  pageSize?: number;
  tenantName?: string;
  tenantCode?: string;
  status?: number;
  channel?: string;
}

export interface TenantListResponse {
  code: string;
  message: string;
  current: number;
  data: TenantListItem[];
  pageSize: number;
  success: boolean;
  total: number;
}

export interface TenantDetailResponse {
  code: string;
  message: string;
  data: TenantListItem;
}

export interface CreateTenantResponse {
  code: string;
  message: string;
  data: CreateTenantResult;
}

export interface TenantStatusPayload {
  ids: number[];
  statusReason?: string;
}

export type TenantStatusAction = 'enable' | 'disable' | 'archive';
