export type ChannelTemplateType = 'channel' | 'integration';
export type ChannelTemplateStatus = 'draft' | 'enabled' | 'disabled' | 'archived';
export type ChannelTemplateScopeType = 'platform' | 'tenant' | 'merchant';

export interface ChannelIntegrationTemplateItem {
  id: number;
  templateCode: string;
  templateName: string;
  templateType: ChannelTemplateType;
  targetCode: string;
  scopeType: ChannelTemplateScopeType;
  platformId: number;
  tenantId: number;
  merchantId: number;
  status: ChannelTemplateStatus;
  metadataConfig: Record<string, any>;
  secretRefConfig: Record<string, any>;
  intentContractConfig: Record<string, any>;
  impactScopeConfig: Record<string, any>;
  remark?: string;
  createBy?: string;
  createTime?: string;
  updateBy?: string;
  updateTime?: string;
  statusReason?: string;
  bindingTenantCount?: number;
  bindingMerchantCount?: number;
  bindingSampleLabels?: string[];
  lastStatusChangeTime?: string;
}

export interface ChannelIntegrationTemplateListParams {
  current?: number;
  pageSize?: number;
  templateName?: string;
  templateType?: ChannelTemplateType;
  targetCode?: string;
  status?: ChannelTemplateStatus;
  tenantId?: number;
  merchantId?: number;
}

export interface ChannelIntegrationTemplateListResponse {
  code: string;
  message: string;
  current: number;
  data: ChannelIntegrationTemplateItem[];
  pageSize: number;
  success: boolean;
  total: number;
}

export interface ChannelIntegrationTemplateDetailResponse {
  code: string;
  message: string;
  data: ChannelIntegrationTemplateItem;
}

export interface ChannelIntegrationTemplateSubmitPayload {
  id?: number;
  templateCode: string;
  templateName: string;
  templateType: ChannelTemplateType;
  targetCode: string;
  scopeType: ChannelTemplateScopeType;
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
  status: ChannelTemplateStatus;
  metadataConfig?: Record<string, any>;
  secretRefConfig?: Record<string, any>;
  intentContractConfig?: Record<string, any>;
  impactScopeConfig?: Record<string, any>;
  remark?: string;
}

export interface ChannelIntegrationTemplateCreateResponse {
  code: string;
  message: string;
  data: {
    id: number;
  };
}

export interface ChannelIntegrationTemplateStatusPayload {
  ids: number[];
  status: ChannelTemplateStatus;
}
