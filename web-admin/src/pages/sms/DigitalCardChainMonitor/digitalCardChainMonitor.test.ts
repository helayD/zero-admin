import {
  buildActionMeta,
  getChainStatusColor,
  normalizeDigitalCardChainListResponse,
} from './helper';

describe('digitalCardChainMonitor helper', () => {
  it('normalizes list response for ProTable consumption', () => {
    const normalized = normalizeDigitalCardChainListResponse({
      data: {
        list: [
          {
            taskId: 11,
            assetNo: 'CARD-001',
          } as any,
        ],
        total: 3,
      },
      success: true,
    });

    expect(normalized.data).toHaveLength(1);
    expect(normalized.total).toBe(3);
    expect(normalized.data[0].taskId).toBe(11);
  });

  it('builds action meta and status colors consistently', () => {
    expect(buildActionMeta('retry')).toMatchObject({
      title: '重试链路',
      successMessage: '链路重试已触发',
    });
    expect(getChainStatusColor('success')).toBe('green');
    expect(getChainStatusColor('failed')).toBe('red');
  });
});
