import React from 'react';
import {
  buildAssetActionMeta,
  getComplianceStatusColor,
  getDisplayStatusColor,
  normalizeDigitalCardAssetListResponse,
} from './helper';
import AssetDetailDrawer from './components/AssetDetailDrawer';

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
    return collectText(node.props?.children);
  }
  return '';
}

describe('digitalCardAssetWorkbench helper', () => {
  it('normalizes list response for ProTable consumption', () => {
    const normalized = normalizeDigitalCardAssetListResponse({
      data: {
        list: [
          {
            assetInstanceId: 11,
            assetNo: 'CARD-001',
          } as any,
        ],
        total: 3,
      },
      success: true,
    });

    expect(normalized.data).toHaveLength(1);
    expect(normalized.total).toBe(3);
    expect(normalized.data[0].assetInstanceId).toBe(11);
  });

  it('builds asset action meta and status colors consistently', () => {
    expect(buildAssetActionMeta('review')).toMatchObject({
      title: '标记人工复核',
      successMessage: '已标记为人工复核',
    });
    expect(getDisplayStatusColor('display_hidden')).toBe('gold');
    expect(getComplianceStatusColor('compliance_recycled')).toBe('red');
  });

  it('renders detail drawer sections and chain monitor entry', () => {
    const element = AssetDetailDrawer({
      open: true,
      onClose: jest.fn(),
      onReview: jest.fn(),
      onOffline: jest.fn(),
      onRecycle: jest.fn(),
      onOpenChainMonitor: jest.fn(),
      detail: {
        item: {
          assetInstanceId: 1,
          assetNo: 'CARD-001',
          memberId: 3001,
          activityId: 2001,
          activityName: '春季抽卡',
          templateId: 21,
          templateName: 'SSR 兔兔',
          cardFaceImage: '',
          rarity: 'SSR',
          tokenId: 'token-1',
          mintTaskId: 11,
          mintStatus: 'mint_processing',
          mintStatusText: '链上处理中',
          chainStatus: 'processing',
          chainStatusText: '处理中',
          displayStatus: 'display_hidden',
          displayStatusText: '受限展示',
          complianceStatus: 'compliance_review',
          complianceStatusText: '人工复核中',
          tokenStatusText: '处理中',
          complianceRuleSummary: '合规复核中',
          lastReceiptSummary: '链上处理中',
          obtainedAt: '2026-04-18 10:00:00',
          disposedAt: '',
          latestReasonSummary: '命中规则待复核',
        },
        participationSummary: {
          participationRecordId: 1,
          requestId: 'req-1',
          resultType: 'won',
          resultStatus: 'won_pending_asset',
          resultStatusText: '已中奖待到账',
          failureReason: '',
          createTime: '2026-04-18 10:00:00',
        },
        mintTaskSummary: {
          taskId: 11,
          requestId: 'req-1',
          traceId: 'trace-1',
          taskStatus: 'running',
          taskStatusText: '处理中',
          mintStatus: 'mint_processing',
          mintStatusText: '链上处理中',
          chainStatus: 'processing',
          chainStatusText: '处理中',
          chainTxId: 'tx-1',
          lastReceiptSummary: '链上处理中',
          availableActions: ['retry', 'freeze'],
        },
        traceId: 'trace-1',
        requestId: 'req-1',
        ruleSnapshotJson: '{"scene":"review"}',
        logs: [
          {
            operationType: 'asset_compliance_review',
            operatorType: 'manual',
            fromStatus: 'compliance_clear',
            toStatus: 'compliance_review',
            reasonText: '命中规则',
            traceId: 'trace-1',
            payloadJson: '{}',
            createTime: '2026-04-18 10:01:00',
          },
        ],
        availableAssetActions: ['review', 'offline', 'recycle'],
      },
    }) as any;
    const text = collectText(element.props.children);

    expect(text).toContain('抽卡来源摘要');
    expect(text).toContain('发放任务摘要');
    expect(text).toContain('完整时间线');
    expect(text).toContain('前往链路监控');
  });
});
