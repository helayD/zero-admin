import type { DigitalCardAssetItem, QueryDigitalCardAssetListResp } from './data';

export const normalizeDigitalCardAssetListResponse = (
  response?: Partial<QueryDigitalCardAssetListResp>,
) => {
  const list = Array.isArray(response?.data?.list) ? response?.data?.list : [];
  const total = Number(response?.data?.total ?? response?.total ?? 0);

  return {
    ...response,
    data: list as DigitalCardAssetItem[],
    total,
    success: response?.success ?? true,
  };
};

export const getChainStatusColor = (status?: string) => {
  switch (status) {
    case 'success':
      return 'green';
    case 'failed':
      return 'red';
    case 'frozen':
      return 'orange';
    case 'processing':
      return 'blue';
    default:
      return 'default';
  }
};

export const getDisplayStatusColor = (status?: string) => {
  switch (status) {
    case 'display_visible':
      return 'green';
    case 'display_hidden':
      return 'gold';
    case 'display_offlined':
      return 'orange';
    case 'display_recycled':
      return 'red';
    default:
      return 'default';
  }
};

export const getComplianceStatusColor = (status?: string) => {
  switch (status) {
    case 'compliance_clear':
      return 'green';
    case 'compliance_review':
      return 'gold';
    case 'compliance_restricted':
      return 'orange';
    case 'compliance_recycle_requested':
      return 'purple';
    case 'compliance_recycled':
      return 'red';
    default:
      return 'default';
  }
};

export const getTimelineColor = (operationType?: string) => {
  switch (operationType) {
    case 'asset_compliance_review':
      return 'gold';
    case 'asset_offlined':
      return 'orange';
    case 'asset_recycled':
      return 'red';
    case 'mint_succeeded':
      return 'green';
    case 'mint_failed':
      return 'red';
    default:
      return 'blue';
  }
};

export const buildAssetActionMeta = (action?: 'review' | 'offline' | 'recycle' | '') => {
  switch (action) {
    case 'review':
      return {
        title: '标记人工复核',
        placeholder: '请填写复核原因，例如：投诉升级、需补充资质核验',
        successMessage: '已标记为人工复核',
        loadingMessage: '正在提交复核处置...',
      };
    case 'offline':
      return {
        title: '下线展示',
        placeholder: '请填写下线原因，例如：内容需临时下架整改',
        successMessage: '资产已下线展示',
        loadingMessage: '正在下线展示...',
      };
    case 'recycle':
      return {
        title: '回收处置',
        placeholder: '请填写回收原因，例如：复核确认违规需平台回收',
        successMessage: '资产已完成回收处置',
        loadingMessage: '正在提交回收处置...',
      };
    default:
      return {
        title: '执行动作',
        placeholder: '请填写原因',
        successMessage: '操作成功',
        loadingMessage: '处理中...',
      };
  }
};
