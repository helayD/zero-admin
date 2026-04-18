import type { QueryDigitalCardChainListResp, DigitalCardChainItem } from './data';

export const toValueEnum = (options: Array<{ label: string; value: string | number }>) =>
  options.reduce<Record<string, { text: string }>>((acc, item) => {
    acc[String(item.value)] = { text: item.label };
    return acc;
  }, {});

export const normalizeDigitalCardChainListResponse = (
  response?: Partial<QueryDigitalCardChainListResp>,
) => {
  const list = Array.isArray(response?.data?.list) ? response?.data?.list : [];
  const total = Number(response?.data?.total ?? response?.total ?? 0);

  return {
    ...response,
    data: list as DigitalCardChainItem[],
    total,
    success: response?.success ?? true,
  };
};

export const getTaskStatusColor = (status?: string) => {
  switch (status) {
    case 'succeeded':
      return 'green';
    case 'failed':
      return 'red';
    case 'manual_review':
      return 'gold';
    case 'frozen':
      return 'orange';
    case 'running':
    case 'dispatched':
      return 'blue';
    default:
      return 'default';
  }
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

export const buildActionMeta = (action?: 'retry' | 'freeze' | 'escalate' | '') => {
  switch (action) {
    case 'retry':
      return {
        title: '重试链路',
        placeholder: '请填写重试原因，例如：链上超时后重新发起补偿',
        successMessage: '链路重试已触发',
        loadingMessage: '正在触发重试...',
      };
    case 'freeze':
      return {
        title: '冻结链路',
        placeholder: '请填写冻结原因，例如：人工核查发现需暂停继续发放',
        successMessage: '链路已冻结',
        loadingMessage: '正在冻结链路...',
      };
    case 'escalate':
      return {
        title: '升级人工复核',
        placeholder: '请填写升级原因，例如：链上返回异常需人工处理',
        successMessage: '链路已升级为人工复核',
        loadingMessage: '正在升级链路...',
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
