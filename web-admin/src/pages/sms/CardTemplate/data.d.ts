export interface CardTemplateListItem {
  id: number; // 模板ID
  templateCode: string; // 模板编码
  templateName: string; // 模板名称
  cardFaceImage: string; // 卡面资源
  copyrightOwner: string; // 版权归属
  copyrightProofSummary: string; // 版权凭证摘要
  rarity: string; // 稀有度
  issueLimit: number; // 发行上限
  displayCopy: string; // 展示文案
  circulationLimitSummary: string; // 默认流转限制
  displayStatus: number; // 展示状态：0-下架，1-上架
  contentAuditStatus: number; // 内容审核状态
  providerCode: string; // 接入方编码
  credentialRef: string; // 实名/版权凭证引用
  status: number; // 启停状态：0-禁用，1-启用
  auditStatus: number; // 审批状态
  platformId: number; // 平台ID
  tenantId: number; // 租户ID
  merchantId: number; // 商户ID
  createBy: string; // 创建人
  updateBy: string; // 更新人
  createTime: string; // 创建时间
  updateTime: string; // 更新时间
  refRuleCount: number; // 引用此模板的发卡规则数量
}

export interface CardTemplateListData {
  list: CardTemplateListItem[];
  total: number;
  page: number;
  pageSize: number;
}

export interface CardTemplateListParams {
  templateName?: string; // 模板名称（模糊查询）
  templateCode?: string; // 模板编码（精确）
  rarity?: string; // 稀有度
  status?: number; // 启停状态：-1-全部，0-禁用，1-启用
  displayStatus?: number; // 展示状态：-1-全部，0-下架，1-上架
  page?: number;
  pageSize?: number;
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
}

export interface AddCardTemplateParams {
  templateCode: string;
  templateName: string;
  cardFaceImage?: string;
  copyrightOwner?: string;
  copyrightProofSummary?: string;
  rarity?: string;
  issueLimit?: number;
  displayCopy?: string;
  circulationLimitSummary?: string;
  displayStatus?: number;
  contentAuditStatus?: number;
  providerCode?: string;
  credentialRef?: string;
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
}

export interface UpdateCardTemplateParams {
  id: number;
  templateName?: string;
  cardFaceImage?: string;
  copyrightOwner?: string;
  copyrightProofSummary?: string;
  rarity?: string;
  issueLimit?: number;
  displayCopy?: string;
  circulationLimitSummary?: string;
  displayStatus?: number;
  contentAuditStatus?: number;
  providerCode?: string;
  credentialRef?: string;
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
}

export interface UpdateCardTemplateStatusParams {
  id: number;
  status: number; // 0-禁用，1-启用
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
}

export interface CheckCardTemplateUsageData {
  refRuleCount: number;
  refRuleNames: string[];
  canDisable: boolean;
  canDelete: boolean;
  message: string;
}

export interface CardTemplateDetailResponse {
  code: string;
  message: string;
  data: CardTemplateListItem;
}

export interface CheckCardTemplateUsageResponse {
  code: string;
  message: string;
  data: CheckCardTemplateUsageData;
}

export const RarityOptions = [
  { value: 'SSR', label: 'SSR' },
  { value: 'SR', label: 'SR' },
  { value: 'R', label: 'R' },
  { value: 'N', label: 'N' },
];

export const ContentAuditStatusOptions = [
  { value: 0, label: '待审' },
  { value: 1, label: '审核中' },
  { value: 2, label: '通过' },
  { value: 3, label: '驳回' },
];
