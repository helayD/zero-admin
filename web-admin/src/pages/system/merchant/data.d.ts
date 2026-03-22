export interface MerchantListItem {
  id: number;
  tenantId: number;
  tenantCode: string;
  tenantName: string;
  tenantStatus: number;
  merchantCode: string;
  merchantName: string;
  merchantShortName?: string;
  contactName: string;
  contactMobile: string;
  contactEmail?: string;
  availableChannels: string[];
  capabilityFlags: string[];
  reviewStatus: number;
  reviewReason?: string;
  reviewedBy: number;
  reviewedByName?: string;
  reviewedAt?: string;
  businessStatus: number;
  statusReason?: string;
  visibleScopeHint?: string;
  primaryAdminUserId?: number;
  remark?: string;
  nextActions: string[];
  createdBy?: string;
  createdAt?: string;
  updatedBy?: string;
  updatedAt?: string;
}

export interface MerchantCreateParams {
  tenantId: number;
  merchantName: string;
  merchantShortName?: string;
  merchantCode?: string;
  contactName: string;
  contactMobile: string;
  contactEmail?: string;
  availableChannels: string[];
  capabilityFlags: string[];
  visibleScopeHint?: string;
  primaryAdminUserId?: number;
  remark?: string;
}

export interface CreateMerchantResult {
  merchantId: number;
  merchantCode: string;
  reviewStatus: number;
  businessStatus: number;
}

export interface MerchantListParams {
  current?: number;
  pageSize?: number;
  tenantId?: number;
  merchantName?: string;
  merchantCode?: string;
  reviewStatus?: number;
  businessStatus?: number;
  channel?: string;
  capabilityFlag?: string;
}

export interface MerchantListResponse {
  code: string;
  message: string;
  current: number;
  data: MerchantListItem[];
  pageSize: number;
  success: boolean;
  total: number;
}

export interface MerchantDetailResponse {
  code: string;
  message: string;
  data: MerchantListItem;
}

export interface CreateMerchantResponse {
  code: string;
  message: string;
  data: CreateMerchantResult;
}

export interface MerchantReviewPayload {
  ids: number[];
  reviewReason?: string;
}

export interface MerchantStatusPayload {
  ids: number[];
  statusReason?: string;
}

export type MerchantReviewAction = 'approve' | 'reject' | 'material';
export type MerchantStatusAction = 'enable' | 'disable' | 'archive';
