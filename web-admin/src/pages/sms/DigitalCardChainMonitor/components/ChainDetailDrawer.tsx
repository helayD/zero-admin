import React from 'react';
import {
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
import { ExclamationCircleOutlined, PauseCircleOutlined, ReloadOutlined } from '@ant-design/icons';
import type { DigitalCardChainDetailData } from '../data';
import { getChainStatusColor, getTaskStatusColor } from '../helper';

const { Paragraph, Text } = Typography;

interface ChainDetailDrawerProps {
  open: boolean;
  loading?: boolean;
  detail?: DigitalCardChainDetailData;
  availableActions?: string[];
  onClose: () => void;
  onRetry: () => void;
  onFreeze: () => void;
  onEscalate: () => void;
}

const ChainDetailDrawer: React.FC<ChainDetailDrawerProps> = ({
  open,
  loading = false,
  detail,
  availableActions = [],
  onClose,
  onRetry,
  onFreeze,
  onEscalate,
}) => {
  const item = detail?.item;

  return (
    <Drawer
      title="数字卡片链路详情"
      placement="right"
      width={720}
      onClose={onClose}
      open={open}
      extra={
        item ? (
          <Space>
            {availableActions.includes('retry') && (
              <Button icon={<ReloadOutlined />} onClick={onRetry}>
                重试
              </Button>
            )}
            {availableActions.includes('freeze') && (
              <Button icon={<PauseCircleOutlined />} onClick={onFreeze}>
                冻结
              </Button>
            )}
            {availableActions.includes('escalate') && (
              <Button danger icon={<ExclamationCircleOutlined />} onClick={onEscalate}>
                升级人工复核
              </Button>
            )}
          </Space>
        ) : null
      }
    >
      {!item ? (
        <Empty description={loading ? '正在加载详情...' : '暂无链路详情'} />
      ) : (
        <Space direction="vertical" size={16} style={{ width: '100%' }}>
          <Descriptions bordered size="small" column={2}>
            <Descriptions.Item label="任务 ID">{item.taskId}</Descriptions.Item>
            <Descriptions.Item label="traceId">
              <Text code>{item.traceId || '-'}</Text>
            </Descriptions.Item>
            <Descriptions.Item label="资产编号">{item.assetNo || '-'}</Descriptions.Item>
            <Descriptions.Item label="资产实例 ID">{item.assetInstanceId}</Descriptions.Item>
            <Descriptions.Item label="会员 ID">{item.memberId}</Descriptions.Item>
            <Descriptions.Item label="中奖记录 ID">{item.participationRecordId}</Descriptions.Item>
            <Descriptions.Item label="活动">
              {item.activityName || `#${item.activityId}`}
            </Descriptions.Item>
            <Descriptions.Item label="模板">
              {item.templateName || `#${item.templateId}`}
            </Descriptions.Item>
            <Descriptions.Item label="任务状态">
              <Tag color={getTaskStatusColor(item.taskStatus)}>
                {item.taskStatusText || item.taskStatus}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="链上状态">
              <Tag color={getChainStatusColor(item.chainStatus)}>
                {item.chainStatusText || item.chainStatus}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="发放状态">
              {item.mintStatusText || item.mintStatus || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="资产阶段">{item.assetStatusText || '-'}</Descriptions.Item>
            <Descriptions.Item label="最近执行时间">{item.lastExecuteAt || '-'}</Descriptions.Item>
            <Descriptions.Item label="重试次数">{item.retryCount}</Descriptions.Item>
            <Descriptions.Item label="人工复核">
              {item.manualRequired ? <Tag color="gold">是</Tag> : <Tag>否</Tag>}
            </Descriptions.Item>
            <Descriptions.Item label="冻结状态">
              {item.frozen ? <Tag color="orange">已冻结</Tag> : <Tag color="green">运行中</Tag>}
            </Descriptions.Item>
            <Descriptions.Item label="冻结原因" span={2}>
              {item.freezeReason || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="tokenId" span={2}>
              {item.tokenId ? <Text code>{item.tokenId}</Text> : '-'}
            </Descriptions.Item>
            <Descriptions.Item label="requestId" span={2}>
              {detail?.requestId ? <Text code>{detail.requestId}</Text> : '-'}
            </Descriptions.Item>
            <Descriptions.Item label="chainTxId" span={2}>
              {detail?.chainTxId ? <Text code>{detail.chainTxId}</Text> : '-'}
            </Descriptions.Item>
            <Descriptions.Item label="最近错误" span={2}>
              {item.lastError ? <Text type="danger">{item.lastError}</Text> : '-'}
            </Descriptions.Item>
            <Descriptions.Item label="最近回执摘要" span={2}>
              {item.lastReceiptSummary || '-'}
            </Descriptions.Item>
          </Descriptions>

          <Divider style={{ margin: 0 }}>最近回执 JSON</Divider>
          {detail?.lastReceiptJson ? (
            <Paragraph copyable={{ text: detail.lastReceiptJson }} style={{ marginBottom: 0 }}>
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
                {detail.lastReceiptJson}
              </pre>
            </Paragraph>
          ) : (
            <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无回执原文" />
          )}

          <Divider style={{ margin: 0 }}>状态时间线</Divider>
          {detail?.logs?.length ? (
            <Timeline>
              {detail.logs.map((log) => (
                <Timeline.Item
                  key={`${log.operationType}-${log.createTime}-${log.traceId}`}
                  color={
                    log.operationType === 'mint_succeeded'
                      ? 'green'
                      : log.operationType === 'mint_failed'
                        ? 'red'
                        : log.operationType === 'mint_manual_review'
                          ? 'gold'
                          : 'blue'
                  }
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

export default ChainDetailDrawer;
