import { request } from 'umi';
import type { OperateHistoryDataListItem, OrderItemDataListItem, OrderListItem, OrderListParams } from './data.d';

type GovernancePayload = {
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
};

const normalizeStatus = (status?: number) => {
  switch (status) {
    case 1:
      return 0;
    case 2:
      return 1;
    case 3:
      return 2;
    case 4:
      return 3;
    case 5:
      return 4;
    default:
      return 5;
  }
};

const toBackendStatus = (status?: number) => {
  switch (status) {
    case 0:
      return 1;
    case 1:
      return 2;
    case 2:
      return 3;
    case 3:
      return 4;
    case 4:
      return 5;
    default:
      return undefined;
  }
};

const normalizeOrderItems = (items?: Record<string, any>[]): OrderItemDataListItem[] =>
  Array.isArray(items)
    ? items.map((item) => ({
        id: item.id,
        orderId: item.orderId,
        orderSn: item.orderNo,
        productId: item.skuId,
        productPic: item.skuPic,
        productName: item.skuName,
        productSn: item.orderNo,
        productPrice: item.skuPrice,
        productQuantity: item.skuQuantity,
        productSkuId: item.skuId,
        productSkuCode: item.skuCode,
        productAttr: item.specData,
        realAmount: item.realAmount,
      }))
    : [];

const normalizeOperateLogs = (logs?: Record<string, any>[]): OperateHistoryDataListItem[] =>
  Array.isArray(logs)
    ? logs.map((log) => ({
        id: log.id,
        orderId: log.orderId,
        operateMan: log.operatorId ? `管理员#${log.operatorId}` : '-',
        createTime: log.createTime,
        note: log.operatorNote,
      }))
    : [];

const normalizeOrder = (item: Record<string, any>): OrderListItem => {
  const deliveryData = item.orderDeliveryData || {};
  const promotions = Array.isArray(item.orderPromotionData) ? item.orderPromotionData : [];

  return {
    id: item.id,
    orderSn: item.orderNo,
    requestTraceId: item.requestTraceId,
    createTime: item.createTime,
    memberUserName: item.userId ? `用户#${item.userId}` : '-',
    totalAmount: item.totalAmount,
    payAmount: item.payAmount,
    freightAmount: item.freightAmount,
    promotionAmount: item.promotionAmount,
    integrationAmount: item.pointsAmount,
    couponAmount: item.couponAmount,
    discountAmount: item.discountAmount,
    payType: item.payType,
    sourceType: item.sourceType,
    status: normalizeStatus(item.orderStatus),
    deliveryCompany: deliveryData.deliveryCompany,
    deliverySn: deliveryData.deliveryNo || item.expressOrderNumber,
    receiverName: deliveryData.receiverName,
    receiverPhone: deliveryData.receiverPhone,
    receiverProvince: deliveryData.receiverProvince,
    receiverCity: deliveryData.receiverCity,
    receiverRegion: deliveryData.receiverDistrict,
    receiverDetailAddress: deliveryData.receiverAddress,
    note: item.remark,
    consistencyStage: item.consistencyStage,
    consistencyStageText: item.consistencyStageText,
    consistencyResult: item.consistencyResult,
    consistencyMessage: item.consistencyMessage,
    lastConsistencyAt: item.lastConsistencyAt,
    pendingActions: item.pendingActions,
    pendingActionsText: item.pendingActionsText,
    aftersaleStatus: item.aftersaleStatus,
    aftersaleStatusText: item.aftersaleStatusText,
    returnId: item.returnId,
    returnNo: item.returnNo,
    listOrderItemData: normalizeOrderItems(item.orderItemData),
    listOperateHistoryData: normalizeOperateLogs(item.orderOperationLogData),
    promotionInfo: promotions.map((promotion) => promotion.promotionName).filter(Boolean).join(' / '),
  };
};

export async function delivery(
  params: GovernancePayload & { orderId: number; deliveryCompany: string; deliverySn: string },
) {
  return request('/api/oms/order/delivery', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function closeOrder(params: GovernancePayload & { ids: number[]; note?: string }) {
  return request('/api/oms/order/closeOrder', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function queryOrderDetail(id: number, scope?: GovernancePayload) {
  const res = await request('/api/oms/order/queryOrderMainDetail', {
    method: 'GET',
    params: {
      id,
      ...scope,
    },
  });

  return {
    ...res,
    data: normalizeOrder(res?.data || {}),
  };
}

export async function queryOrderList(params: OrderListParams) {
  const { status, ...rest } = params;
  const res = await request('/api/oms/order/queryOrderMainList', {
    method: 'GET',
    params: {
      ...rest,
      orderStatus: toBackendStatus(status),
    },
  });

  return {
    ...res,
    data: Array.isArray(res?.data) ? res.data.map(normalizeOrder) : [],
  };
}

export async function updateMoneyInfo(
  params: GovernancePayload & { id: number; status?: number; freightAmount: number; discountAmount: number },
) {
  return request('/api/oms/order/updateMoneyInfo', {
    method: 'POST',
    data: {
      ...params,
      status: toBackendStatus(params.status),
    },
  });
}

export async function updateNote(
  params: GovernancePayload & { id: number; status?: number; note: string },
) {
  return request('/api/oms/order/updateNote', {
    method: 'POST',
    data: {
      ...params,
      status: toBackendStatus(params.status),
    },
  });
}

export async function removeOrder(ids: number[], scope?: GovernancePayload) {
  return request('/api/oms/order/deleteOrderMain', {
    method: 'GET',
    params: {
      ids: ids.join(','),
      ...scope,
    },
  });
}
