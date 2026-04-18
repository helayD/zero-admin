import React from 'react';
import {
  Alert,
  Button,
  Descriptions,
  Divider,
  Drawer,
  Empty,
  Space,
  Tag,
  Timeline,
  Typography,
} from 'antd';
import {
  ExclamationCircleOutlined,
  EyeInvisibleOutlined,
  SafetyCertificateOutlined,
  SendOutlined,
} from '@ant-design/icons';
import type { DigitalCardAssetDetailData } from '../data';
import {
  getChainStatusColor,
  getComplianceStatusColor,
  getDisplayStatusColor,
  getTimelineColor,
} from '../helper';

const { Paragraph, Text } = Typography;

interface AssetDetailDrawerProps {
  open: boolean;
  loading?: boolean;
  detail?: DigitalCardAssetDetailData;
  onClose: () => void;
  onReview: () => void;
  onOffline: () => void;
  onRecycle: () => void;
  onOpenChainMonitor: () => void;
}

const AssetDetailDrawer: React.FC<AssetDetailDrawerProps> = ({
  open,
  loading = false,
  detail,
  onClose,
  onReview,
  onOffline,
  onRecycle,
  onOpenChainMonitor,
}) => {
  const item = detail?.item;
  const assetActions = detail?.availableAssetActions || [];
  const chainActions = detail?.mintTaskSummary?.availableActions || [];

  return (
    <Drawer
      title="数字卡片资产详情"
      placement="right"
      width={760}
      onClose={onClose}
      open={open}
      extra={
        item ? (
          <Space wrap>
            {assetActions.includes('review') && (
              <Button icon={<SafetyCertificateOutlined />} onClick={onReview}>
                标记复核
              </Button>
            )}
            {assetActions.includes('offline') && (
              <Button icon={<EyeInvisibleOutlined />} onClick={onOffline}>
                下线展示
              </Button>
            )}
            {assetActions.includes('recycle') && (
              <Button danger icon={<ExclamationCircleOutlined />} onClick={onRecycle}>
                回收处置
              </Button>
            )}
          </Space>
        ) : null
      }
    >
      {!item ? (
        <Empty description={loading ? '正在加载资产详情...' : '暂无资产详情'} />
      ) : (
        <Space direction="vertical" size={16} style={{ width: '100%' }}>
          <Alert
            type="info"
            showIcon
            message="资产级动作与链路级动作分开展示"
            description="review / offline / recycle 会直接变更资产展示与合规状态；retry / freeze / escalate 继续前往链路监控页处理。"
          />

          <Descriptions bordered size="small" column={2}>
            <Descriptions.Item label="资产实例 ID">{item.assetInstanceId}</Descriptions.Item>
            <Descriptions.Item label="资产编号">{item.assetNo || '-'}</Descriptions.Item>
            <Descriptions.Item label="会员 ID">{item.memberId}</Descriptions.Item>
            <Descriptions.Item label="发放任务 ID">{item.mintTaskId || '-'}</Descriptions.Item>
            <Descriptions.Item label="活动">
              {item.activityName || `#${item.activityId}`}
            </Descriptions.Item>
            <Descriptions.Item label="模板">
              {item.templateName || `#${item.templateId}`}
            </Descriptions.Item>
            <Descriptions.Item label="链上状态">
              <Tag color={getChainStatusColor(item.chainStatus)}>
                {item.chainStatusText || item.chainStatus || '-'}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="发放状态">
              {item.mintStatusText || item.mintStatus || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="展示状态">
              <Tag color={getDisplayStatusColor(item.displayStatus)}>
                {item.displayStatusText || item.displayStatus || '-'}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="合规状态">
              <Tag color={getComplianceStatusColor(item.complianceStatus)}>
                {item.complianceStatusText || item.complianceStatus || '-'}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="token 状态">{item.tokenStatusText || '-'}</Descriptions.Item>
            <Descriptions.Item label="最近处置时间">{item.disposedAt || '-'}</Descriptions.Item>
            <Descriptions.Item label="获取时间">{item.obtainedAt || '-'}</Descriptions.Item>
            <Descriptions.Item label="tokenId" span={2}>
              {item.tokenId ? <Text code>{item.tokenId}</Text> : '-'}
            </Descriptions.Item>
            <Descriptions.Item label="最近回执摘要" span={2}>
              {item.lastReceiptSummary || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="最近原因摘要" span={2}>
              {item.latestReasonSummary || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="合规提示" span={2}>
              {item.complianceRuleSummary || '-'}
            </Descriptions.Item>
          </Descriptions>

          <Divider style={{ margin: 0 }}>抽卡来源摘要</Divider>
          <Descriptions bordered size="small" column={2}>
            <Descriptions.Item label="抽卡记录 ID">
              {detail?.participationSummary?.participationRecordId || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="参与时间">
              {detail?.participationSummary?.createTime || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="结果状态">
              {detail?.participationSummary?.resultStatusText ||
                detail?.participationSummary?.resultStatus ||
                '-'}
            </Descriptions.Item>
            <Descriptions.Item label="失败原因">
              {detail?.participationSummary?.failureReason || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="requestId" span={2}>
              {detail?.participationSummary?.requestId ? (
                <Text code>{detail?.participationSummary?.requestId}</Text>
              ) : (
                '-'
              )}
            </Descriptions.Item>
          </Descriptions>

          <Divider style={{ margin: 0 }}>发放任务摘要</Divider>
          <Space direction="vertical" size={12} style={{ width: '100%' }}>
            <Descriptions bordered size="small" column={2}>
              <Descriptions.Item label="任务状态">
                {detail?.mintTaskSummary?.taskStatusText ||
                  detail?.mintTaskSummary?.taskStatus ||
                  '-'}
              </Descriptions.Item>
              <Descriptions.Item label="链上交易 ID">
                {detail?.mintTaskSummary?.chainTxId ? (
                  <Text code>{detail?.mintTaskSummary?.chainTxId}</Text>
                ) : (
                  '-'
                )}
              </Descriptions.Item>
              <Descriptions.Item label="traceId" span={2}>
                {detail?.mintTaskSummary?.traceId ? (
                  <Text code>{detail?.mintTaskSummary?.traceId}</Text>
                ) : (
                  '-'
                )}
              </Descriptions.Item>
              <Descriptions.Item label="最近回执摘要" span={2}>
                {detail?.mintTaskSummary?.lastReceiptSummary || '-'}
              </Descriptions.Item>
            </Descriptions>

            <Space wrap>
              <Button type="primary" icon={<SendOutlined />} onClick={onOpenChainMonitor}>
                前往链路监控
              </Button>
              {chainActions.map((action) => (
                <Tag key={action} color="blue">
                  {action}
                </Tag>
              ))}
            </Space>
          </Space>

          <Divider style={{ margin: 0 }}>合规快照</Divider>
          {detail?.ruleSnapshotJson ? (
            <Paragraph copyable={{ text: detail.ruleSnapshotJson }} style={{ marginBottom: 0 }}>
              <pre
                style={{
                  margin: 0,
                  padding: 12,
                  background: '#fafafa',
                  borderRadius: 6,
                  whiteSpace: 'pre-wrap',
                  wordBreak: 'break-all',
                }}
              >
                {detail.ruleSnapshotJson}
              </pre>
            </Paragraph>
          ) : (
            <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无合规快照" />
          )}

          <Divider style={{ margin: 0 }}>完整时间线</Divider>
          {detail?.logs?.length ? (
            <Timeline>
              {detail.logs.map((log) => (
                <Timeline.Item
                  key={`${log.operationType}-${log.createTime}-${log.traceId}`}
                  color={getTimelineColor(log.operationType)}
                >
                  <Space direction="vertical" size={4}>
                    <Space wrap>
                      <Text strong>{log.operationType}</Text>
                      <Tag>{log.operatorType || '-'}</Tag>
                      <Text type="secondary">{log.createTime || '-'}</Text>
                    </Space>
                    <Text>
                      {log.fromStatus || '-'} → {log.toStatus || '-'}
                    </Text>
                    <Text type="secondary">{log.reasonText || '无附加说明'}</Text>
                    {log.traceId ? <Text code>{log.traceId}</Text> : null}
                  </Space>
                </Timeline.Item>
              ))}
            </Timeline>
          ) : (
            <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无时间线日志" />
          )}
        </Space>
      )}
    </Drawer>
  );
};

export default AssetDetailDrawer;
