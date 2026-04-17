export type GovernanceScopeType = 'platform' | 'tenant' | 'merchant';

export interface DrawReadinessItem {
  code: string;
  field: string;
  message: string;
  blocking: boolean;
}

export interface DrawHomeEntryConfig {
  showOnHome?: number;
  homeEntryTitle?: string;
  homeEntrySubtitle?: string;
  homeEntryImage?: string;
  homeEntrySort?: number;
  landingTargetType?: string;
  landingTargetValue?: string;
  isEnabled?: number;
}

export interface DrawCardTemplateData {
  id?: number;
  templateName: string;
  templateCode: string;
  cardFaceImage?: string;
  copyrightOwner?: string;
  copyrightProofSummary?: string;
  rarity?: string;
  issueLimit: number;
  displayCopy?: string;
  circulationLimitSummary?: string;
  displayStatus?: number;
  contentAuditStatus?: number;
  providerCode?: string;
  credentialRef?: string;
  status?: number;
  auditStatus?: number;
}

export interface DrawPoolTemplateData {
  id?: number;
  templateId?: number;
  templateCode?: string;
  templateName?: string;
  rarity?: string;
  probability: number;
  saleLimit: number;
  remainingLimit: number;
  configLimit: number;
}

export interface DrawPoolData {
  id?: number;
  poolName: string;
  poolCode: string;
  probabilityRule?: string;
  sort?: number;
  status?: number;
  templates?: DrawPoolTemplateData[];
}

export interface DrawActivityListItem {
  id?: number;
  activityCode: string;
  name: string;
  ruleSummary?: string;
  startTime: string;
  endTime: string;
  realNameRequired?: number;
  participantConditionSummary?: string;
  consumeRuleSummary?: string;
  probabilityRule?: string;
  complianceRuleSummary?: string;
  circulationLimitSummary?: string;
  approvalRecordRef?: string;
  copyrightStatus?: number;
  contentAuditStatus?: number;
  status?: number;
  auditStatus?: number;
  isEnabled?: number;
  publishReadiness?: number;
  publishFailureSummary?: string;
  scopeType?: GovernanceScopeType;
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
  updateTime?: string;
  createTime?: string;
  createBy?: number;
  updateBy?: number;
  showOnHome?: number;
  homeEntryTitle?: string;
  readinessLabel?: string;
  homeEntry?: DrawHomeEntryConfig;
  templates?: DrawCardTemplateData[];
  pools?: DrawPoolData[];
  readinessItems?: DrawReadinessItem[];
  audits?: {
    id: number;
    operationType: string;
    operatorName: string;
    approvalResult: string;
    failureSummary: string;
    traceId: string;
    createTime: string;
  }[];
}

export interface DrawActivityListParams {
  current?: number;
  pageSize?: number;
  name?: string;
  status?: number;
  auditStatus?: number;
  publishReadiness?: number;
  scopeType?: GovernanceScopeType;
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
}

export interface PreviewDrawActivityPublishReadinessData {
  readyToPublish: boolean;
  publishReadiness: number;
  readinessLabel: string;
  summary: string;
  items: DrawReadinessItem[];
}

export interface DrawActivityFormValues extends DrawActivityListItem {
  activeTime?: any;
}
