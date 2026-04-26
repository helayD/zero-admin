import type {
  DigitalCardPhysicalFulfillmentItem,
  QueryDigitalCardPhysicalFulfillmentListResp,
} from './data';

export const normalizePhysicalFulfillmentListResponse = (
  response?: Partial<QueryDigitalCardPhysicalFulfillmentListResp>,
) => {
  const list = Array.isArray(response?.data?.list) ? response?.data?.list : [];
  const total = Number(response?.data?.total ?? response?.total ?? 0);

  return {
    ...response,
    data: list as DigitalCardPhysicalFulfillmentItem[],
    total,
    success: response?.success ?? true,
  };
};

export const getFulfillmentStatusColor = (status?: string) => {
  switch (status) {
    case 'signed':
      return 'green';
    case 'shipped':
    case 'in_transit':
      return 'blue';
    case 'ready_to_ship':
    case 'production_pending':
    case 'production_in_progress':
    case 'quality_checking':
      return 'gold';
    case 'exception':
    case 'cancelled':
      return 'red';
    case 'reissue_pending':
      return 'purple';
    default:
      return 'default';
  }
};

export const getTimelineColor = (action?: string) => {
  switch (action) {
    case 'physical_signed':
      return 'green';
    case 'physical_shipped':
      return 'blue';
    case 'physical_exception':
      return 'red';
    case 'physical_reissue_requested':
      return 'purple';
    case 'physical_production_updated':
      return 'gold';
    default:
      return 'blue';
  }
};

export const buildPhysicalActionMeta = (
  action?: 'ensure' | 'production' | 'ship' | 'exception' | 'reissue' | '',
) => {
  switch (action) {
    case 'ensure':
      return {
        title: '初始化履约单',
        successMessage: '履约单已初始化',
        loadingMessage: '正在初始化履约单...',
      };
    case 'production':
      return {
        title: '更新制作状态',
        successMessage: '制作状态已更新',
        loadingMessage: '正在更新制作状态...',
      };
    case 'ship':
      return {
        title: '登记物流发货',
        successMessage: '物流发货已登记',
        loadingMessage: '正在登记物流...',
      };
    case 'exception':
      return {
        title: '标记履约异常',
        successMessage: '履约异常已标记',
        loadingMessage: '正在标记异常...',
      };
    case 'reissue':
      return {
        title: '发起实体卡补发',
        successMessage: '实体卡补发已发起',
        loadingMessage: '正在发起补发...',
      };
    default:
      return {
        title: '实体卡履约动作',
        successMessage: '操作成功',
        loadingMessage: '处理中...',
      };
  }
};
