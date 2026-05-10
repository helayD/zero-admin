import { request } from 'umi';
import moment from 'moment';
import type {
  CategoryListParams,
  CouponHistoryListParams,
  CouponListItem,
  CouponListParams,
} from './data.d';
import type { ProductListParams } from '@/pages/pms/product/data';

type GovernancePayload = {
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
};

type CouponScopeRow = {
  id?: number;
  name?: string;
  productSn?: string;
};

type CouponFormPayload = CouponListItem & GovernancePayload;

const formatDateTime = (value?: moment.MomentInput) => {
  if (!value) {
    return '';
  }
  return moment(value).format('YYYY-MM-DD HH:mm:ss');
};

const buildCouponScopeData = (params: CouponListItem) => {
  if (params.useType === 1) {
    return (params.productCategoryRelationList || []).map((item: CouponScopeRow) => ({
      scopeType: 1,
      scopeId: item.id,
    }));
  }
  if (params.useType === 2) {
    return (params.productRelationList || []).map((item: CouponScopeRow) => ({
      scopeType: 2,
      scopeId: item.id,
    }));
  }
  return [];
};

const normalizeCouponItem = (item: Record<string, any>) => ({
  ...item,
  type: item.type ?? item.typeId ?? 0,
  platform: item.platform ?? 0,
  useType: item.useType ?? item.scopeType ?? 0,
  minPoint: item.minPoint ?? item.minAmount ?? 0,
  note: item.note ?? item.description ?? '',
  publishCount: item.publishCount ?? item.totalCount ?? 0,
  useCount: item.useCount ?? item.usedCount ?? 0,
  receiveCount: item.receiveCount ?? item.receivedCount ?? 0,
  count: item.count ?? item.totalCount ?? 0,
  memberLevel: item.memberLevel ?? 0,
  enableTime: item.enableTime ?? item.startTime ?? '',
  productCategoryRelationList:
    item.productCategoryRelationList ||
    ((item.scopeType ?? 0) === 1 ? item.couponScopeData || [] : []),
  productRelationList:
    item.productRelationList ||
    ((item.scopeType ?? 0) === 2 ? item.couponScopeData || [] : []),
});

export async function queryCoupon(params: CouponListParams) {
  const { type, ...rest } = params;
  const res = await request('/api/sms/coupon/queryCouponList', {
    method: 'GET',
    params: {
      ...rest,
      typeId: type != null ? Number(type) : undefined,
    },
  });

  return {
    ...res,
    data: Array.isArray(res?.data) ? res.data.map(normalizeCouponItem) : [],
  };
}

export async function queryCouponDetail(id: number, scope?: GovernancePayload) {
  const res = await request('/api/sms/coupon/queryCouponDetail', {
    method: 'GET',
    params: {
      id,
      ...scope,
    },
  });

  return {
    ...res,
    data: normalizeCouponItem(res?.data || {}),
  };
}

export async function removeCoupon(params: { ids: number[] } & GovernancePayload) {
  return request('/api/sms/coupon/deleteCoupon', {
    method: 'GET',
    params: {
      ...params,
    },
  });
}

export async function addCoupon(params: CouponFormPayload) {
  return request('/api/sms/coupon/addCoupon', {
    method: 'POST',
    data: {
      typeId: Number(params.typeId || params.type || 0),
      name: params.name,
      code: params.code || '',
      amount: Number(params.amount || 0),
      minAmount: Number(params.minAmount || params.minPoint || 0),
      startTime: formatDateTime(params.startTime?.[0]),
      endTime: formatDateTime(params.startTime?.[1]),
      totalCount: Number(params.totalCount || params.publishCount || 0),
      perLimit: Number(params.perLimit || 0),
      isEnabled: params.isEnabled ?? 1,
      description: params.description || params.note || '',
      couponScopeData: buildCouponScopeData(params),
      scopeType: params.scopeType,
      platformId: params.platformId,
      tenantId: params.tenantId,
      merchantId: params.merchantId,
    },
  });
}

export async function updateCoupon(params: CouponFormPayload) {
  return request('/api/sms/coupon/updateCoupon', {
    method: 'POST',
    data: {
      id: params.id,
      typeId: Number(params.typeId || params.type || 0),
      name: params.name,
      code: params.code || '',
      amount: Number(params.amount || 0),
      minAmount: Number(params.minAmount || params.minPoint || 0),
      startTime: formatDateTime(params.startTime?.[0]),
      endTime: formatDateTime(params.startTime?.[1]),
      totalCount: Number(params.totalCount || params.publishCount || 0),
      receivedCount: Number(params.receivedCount || params.receiveCount || 0),
      usedCount: Number(params.usedCount || params.useCount || 0),
      perLimit: Number(params.perLimit || 0),
      isEnabled: params.isEnabled ?? 1,
      description: params.description || params.note || '',
      couponScopeData: buildCouponScopeData(params),
      scopeType: params.scopeType,
      platformId: params.platformId,
      tenantId: params.tenantId,
      merchantId: params.merchantId,
    },
  });
}

export async function updateCouponStatus(params: { ids: number[]; status: number } & GovernancePayload) {
  return request('/api/sms/coupon/updateCouponStatus', {
    method: 'POST',
    data: params,
  });
}

export async function queryProductList(params?: ProductListParams & GovernancePayload) {
  return request('/api/pms/product/queryProductSpuList', {
    method: 'GET',
    params: {
      ...params,
    },
  }).then((res) => ({
    ...res,
    data: Array.isArray(res?.data)
      ? res.data.map((item: Record<string, any>) => ({
          id: item.id,
          name: item.name,
          productSn: item.productSn,
          promotionPrice: item.priceRange,
          originalPrice: item.priceRange,
          stock: item.stock,
        }))
      : [],
  }));
}

export async function queryProductCategoryList(params: CategoryListParams & GovernancePayload) {
  return request('/api/pms/product/category/queryProductCategoryList', {
    method: 'GET',
    params: {
      ...params,
    },
  });
}

export async function queryCouponHistoryList(params: CouponHistoryListParams) {
  const res = await request('/api/sms/couponRecord/queryCouponRecordList', {
    method: 'GET',
    params: {
      ...params,
    },
  });

  return {
    ...res,
    data: Array.isArray(res?.data)
      ? res.data.map((item: Record<string, any>) => ({
          ...item,
          // 后端返回 status（0-未使用/1-已使用/2-已过期/3-已失效），
          // 前端 CouponHistoryListItem 字段名为 useStatus，做一次兼容映射
          useStatus: item.useStatus ?? item.status,
        }))
      : [],
  };
}
