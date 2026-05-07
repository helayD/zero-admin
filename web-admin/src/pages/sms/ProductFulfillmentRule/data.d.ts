export interface ProductFulfillmentRuleListItem {
  id: number; // 规则ID
  ruleName: string; // 规则名称
  ruleStatus: number; // 规则状态：0-禁用，1-启用
  cardTemplateId: number; // 关联卡片模板ID
  cardTemplateName: string; // 关联卡片模板名称
  expireDays: number; // 卡片有效期天数
  transferable: number; // 是否可转赠：0-否，1-是
  transferLimit: number; // 最大转赠次数
  claimCondition: string; // 领取条件（JSON）
  redemptionCondition: string; // 提货条件
  refundPolicy: string; // 退款处置策略
  refundPolicyText: string; // 退款处置策略文案
  platformId: number; // 平台ID
  tenantId: number; // 租户ID
  merchantId: number; // 商户ID
  createBy: string; // 创建人
  updateBy: string; // 更新人
  createTime: string; // 创建时间
  updateTime: string; // 更新时间
  bindingCount: number; // 绑定商品数量
  scopeType?: 'platform' | 'tenant' | 'merchant'; // 作用域来源
}

export interface ProductFulfillmentRuleListData {
  list: ProductFulfillmentRuleListItem[];
  total: number;
  page: number;
  pageSize: number;
}

export interface ProductFulfillmentRuleListParams {
  ruleName?: string; // 规则名称（模糊查询）
  cardTemplateId?: number; // 关联卡片模板ID
  ruleStatus?: number; // 规则状态：-1-全部，0-禁用，1-启用
  page?: number; // 页码
  pageSize?: number; // 每页数量
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
}

export interface AddProductFulfillmentRuleParams {
  ruleName: string; // 规则名称
  cardTemplateId: number; // 关联卡片模板ID
  expireDays: number; // 卡片有效期天数
  transferable: number; // 是否可转赠：0-否，1-是
  transferLimit: number; // 最大转赠次数
  claimCondition?: string; // 领取条件（JSON）
  redemptionCondition?: string; // 提货条件
  refundPolicy: string; // 退款处置策略
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
}

export interface UpdateProductFulfillmentRuleParams {
  id: number; // 规则ID
  ruleName?: string; // 规则名称
  cardTemplateId?: number; // 关联卡片模板ID
  expireDays?: number; // 卡片有效期天数
  transferable?: number; // 是否可转赠：0-否，1-是
  transferLimit?: number; // 最大转赠次数
  claimCondition?: string; // 领取条件（JSON）
  redemptionCondition?: string; // 提货条件
  refundPolicy?: string; // 退款处置策略
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
}

export interface UpdateProductFulfillmentRuleStatusParams {
  id: number; // 规则ID
  ruleStatus: number; // 规则状态：0-禁用，1-启用
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
}

export interface CheckProductFulfillmentRuleBindingData {
  bindingCount: number; // 绑定商品数量
  bindingProductNames: string[]; // 绑定的商品名称列表
  canDisable: boolean; // 是否可以禁用
  canDelete: boolean; // 是否可以删除
  message: string; // 提示信息
}

export interface ProductFulfillmentRuleDetailResponse {
  code: string;
  message: string;
  data: ProductFulfillmentRuleListItem;
}

export interface CheckProductFulfillmentRuleBindingResponse {
  code: string;
  message: string;
  data: CheckProductFulfillmentRuleBindingData;
}

export const RefundPolicyOptions = [
  { value: 'freeze_card', label: '冻结卡片' },
  { value: 'recycle_card', label: '回收卡片' },
  { value: 'manual_review', label: '人工复核' },
];

export const RefundPolicyTextMap: Record<string, string> = {
  freeze_card: '冻结卡片',
  recycle_card: '回收卡片',
  manual_review: '人工复核',
};
