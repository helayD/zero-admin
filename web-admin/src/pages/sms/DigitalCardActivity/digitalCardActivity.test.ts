import moment from 'moment';
import { buildPreviewMessages, serializeDrawActivityPayload } from './helper';

describe('digitalCardActivity helper', () => {
  it('serializes date range and nested pools for request payloads', () => {
    const payload = serializeDrawActivityPayload({
      activityCode: 'DRAW20260416',
      name: '春季抽卡活动',
      activeTime: [moment('2026-04-16 10:00:00'), moment('2026-04-30 23:59:59')],
      homeEntry: {
        showOnHome: 1,
        homeEntryTitle: '首页抽卡',
      },
      templates: [
        {
          templateName: 'SSR 兔兔',
          templateCode: 'T-SSR-01',
          issueLimit: 100,
        },
      ],
      pools: [
        {
          poolName: '新手池',
          poolCode: 'POOL-01',
          templates: [
            {
              templateCode: 'T-SSR-01',
              probability: 1,
              saleLimit: 100,
              remainingLimit: 100,
              configLimit: 100,
            },
          ],
        },
      ],
    } as any);

    expect(payload.startTime).toBe('2026-04-16 10:00:00');
    expect(payload.endTime).toBe('2026-04-30 23:59:59');
    expect(payload.templates).toHaveLength(1);
    expect(payload.pools[0].templates[0].templateCode).toBe('T-SSR-01');
  });

  it('builds readable preview messages for blocking and non-blocking items', () => {
    expect(
      buildPreviewMessages({
        readyToPublish: false,
        publishReadiness: 0,
        readinessLabel: '缺少信息',
        summary: '审批记录缺失',
        items: [
          {
            code: 'missingAuditRecord',
            field: 'approvalRecordRef',
            message: '审批记录缺失',
            blocking: true,
          },
          {
            code: 'pendingContentAudit',
            field: 'contentAuditStatus',
            message: '内容审核处理中',
            blocking: false,
          },
        ],
      }),
    ).toEqual(['阻断: 审批记录缺失', '待处理: 内容审核处理中']);
  });
});
