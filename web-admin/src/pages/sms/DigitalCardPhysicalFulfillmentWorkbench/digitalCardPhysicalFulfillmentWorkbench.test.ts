import React from 'react';
import {
  buildPhysicalActionMeta,
  getFulfillmentStatusColor,
  normalizePhysicalFulfillmentListResponse,
} from './helper';
import PhysicalFulfillmentDetailDrawer from './components/PhysicalFulfillmentDetailDrawer';

function collectText(node: any): string {
  if (node === null || node === undefined || typeof node === 'boolean') {
    return '';
  }
  if (typeof node === 'string' || typeof node === 'number') {
    return String(node);
  }
  if (Array.isArray(node)) {
    return node.map((item) => collectText(item)).join(' ');
  }
  if (React.isValidElement(node)) {
    return collectText((node.props as any)?.children);
  }
  return '';
}

describe('digitalCardPhysicalFulfillmentWorkbench helper', () => {
  it('normalizes list response for ProTable consumption', () => {
    const normalized = normalizePhysicalFulfillmentListResponse({
      data: {
        list: [
          {
            fulfillmentId: 9,
            fulfillmentNo: 'PF2026042600000009',
          } as any,
        ],
        total: 2,
      },
      success: true,
    });

    expect(normalized.data).toHaveLength(1);
    expect(normalized.total).toBe(2);
    expect(normalized.data[0].fulfillmentId).toBe(9);
  });

  it('builds action meta and status colors consistently', () => {
    expect(buildPhysicalActionMeta('ship')).toMatchObject({
      title: '登记物流发货',
      successMessage: '物流发货已登记',
    });
    expect(getFulfillmentStatusColor('signed')).toBe('green');
    expect(getFulfillmentStatusColor('exception')).toBe('red');
  });

  it('renders physical fulfillment detail sections and actions', () => {
    const element = PhysicalFulfillmentDetailDrawer({
      open: true,
      onClose: jest.fn(),
      onUpdateProduction: jest.fn(),
      onShip: jest.fn(),
      onException: jest.fn(),
      onReissue: jest.fn(),
      detail: {
        item: {
          fulfillmentId: 9,
          fulfillmentNo: 'PF2026042600000009',
          assetInstanceId: 1,
          assetNo: 'CARD-001',
          activityId: 2001,
          activityName: '春季抽卡',
          templateId: 21,
          templateName: '实体卡模板',
          memberId: 3001,
          fulfillmentStatus: 'shipped',
          fulfillmentStatusText: '已发货',
          productionStatus: 'completed',
          productionStatusText: '制作完成',
          shippingStatus: 'shipped',
          shippingStatusText: '已发货',
          productionBatchNo: 'BATCH-001',
          carrierName: '顺丰速运',
          trackingNo: 'SF10001',
          failureCode: '',
          failureReason: '',
          shippedAt: '2026-04-26 10:00:00',
          signedAt: '',
          updateTime: '2026-04-26 10:00:00',
        },
        receiverName: '张三',
        receiverPhone: '13800138000',
        addressSummary: '浙江省杭州市西湖区文三路 1 号',
        requestId: 'req-1',
        traceId: 'trace-1',
        timeline: [
          {
            action: 'physical_shipped',
            actionText: '实体卡发货',
            statusText: '已发货',
            reason: '批量发货',
            createTime: '2026-04-26 10:00:00',
          },
        ],
      },
    }) as any;
    const text = collectText(element.props.children);
    const actionText = collectText(element.props.extra);

    expect(text).toContain('PF2026042600000009');
    expect(text).toContain('浙江省杭州市西湖区文三路 1 号');
    expect(text).toContain('实体卡发货');
    expect(actionText).toContain('登记物流');
  });
});
