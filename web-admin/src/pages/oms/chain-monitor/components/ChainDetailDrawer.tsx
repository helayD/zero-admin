import React from 'react';
import { Button, Descriptions, Divider, Drawer, Space, Tag } from 'antd';
import {
  ExclamationCircleOutlined,
  PauseCircleOutlined,
  PlayCircleOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import type { ChainMonitorItem } from '../data.d';

interface ChainDetailDrawerProps {
  visible: boolean;
  chainData?: ChainMonitorItem;
  availableActions?: string[];
  onClose: () => void;
  onRetry: (orderId: number) => void;
  onReplay: (orderId: number) => void;
  onPause: (orderId: number) => void;
  onEscalate: (orderId: number) => void;
}

const ChainDetailDrawer: React.FC<ChainDetailDrawerProps> = ({
  visible,
  chainData,
  availableActions = [],
  onClose,
  onRetry,
  onReplay,
  onPause,
  onEscalate,
}) => {
  if (!chainData) return null;

  const isRetryAvailable = availableActions.includes('retry');
  const isReplayAvailable = availableActions.includes('replay');
  const isPauseAvailable = availableActions.includes('pause');
  const isEscalateAvailable = availableActions.includes('escalate');

  return (
    <Drawer
      title="链路详情"
      placement="right"
      width={620}
      onClose={onClose}
      open={visible}
      extra={
        <Space>
          {isRetryAvailable && (
            <Button
              icon={<ReloadOutlined />}
              onClick={() => onRetry(chainData.entityId)}
            >
              重试
            </Button>
          )}
          {isReplayAvailable && (
            <Button
              icon={<PlayCircleOutlined />}
              onClick={() => onReplay(chainData.entityId)}
            >
              回放
            </Button>
          )}
          {isPauseAvailable && (
            <Button
              icon={<PauseCircleOutlined />}
              onClick={() => onPause(chainData.entityId)}
            >
              暂停
            </Button>
          )}
          {isEscalateAvailable && (
            <Button
              icon={<ExclamationCircleOutlined />}
              danger
              onClick={() => onEscalate(chainData.entityId)}
            >
              升级
            </Button>
          )}
        </Space>
      }
    >
      <Descriptions column={2} bordered size="small">
        <Descriptions.Item label="链路追踪ID" span={2}>
          <span style={{ fontFamily: 'monospace' }}>{chainData.traceId}</span>
        </Descriptions.Item>

        <Descriptions.Item label="平台ID">{chainData.platformId}</Descriptions.Item>
        <Descriptions.Item label="租户ID">{chainData.tenantId}</Descriptions.Item>

        <Descriptions.Item label="链路类型">{chainData.chainTypeText}</Descriptions.Item>
        <Descriptions.Item label="业务对象类型">
          {chainData.entityType === 1 ? '订单' : chainData.entityType === 2 ? '商品' : '-'}
        </Descriptions.Item>

        <Descriptions.Item label="业务编号">{chainData.entityNo}</Descriptions.Item>
        <Descriptions.Item label="业务ID">{chainData.entityId}</Descriptions.Item>

        <Descriptions.Item label="当前阶段">{chainData.stageText}</Descriptions.Item>
        <Descriptions.Item label="处理结果">
          <Tag color={chainData.resultText === '失败' || chainData.resultText === '人工处理' ? 'red' : chainData.resultText === '处理中' ? 'blue' : 'green'}>
            {chainData.resultText}
          </Tag>
        </Descriptions.Item>

        <Descriptions.Item label="重试次数">
          {chainData.retryCount >= 3 ? (
            <span style={{ color: '#ff4d4f', fontWeight: 600 }}>{chainData.retryCount} ⚠️</span>
          ) : (
            chainData.retryCount
          )}
        </Descriptions.Item>
        <Descriptions.Item label="触发操作人ID">{chainData.actorId || '-'}</Descriptions.Item>

        <Descriptions.Item label="暂停状态" span={2}>
          {chainData.paused === 1 ? (
            <Tag color="orange">已暂停{chainData.pauseReason ? `：${chainData.pauseReason}` : ''}</Tag>
          ) : (
            <Tag color="green">运行中</Tag>
          )}
        </Descriptions.Item>

        <Descriptions.Item label="最近执行时间" span={2}>
          {chainData.lastExecuteAt || '-'}
        </Descriptions.Item>

        <Descriptions.Item label="链路创建时间" span={2}>
          {chainData.createdAt || '-'}
        </Descriptions.Item>

        <Descriptions.Item label="最近错误" span={2}>
          {chainData.lastError ? (
            <span style={{ color: '#ff4d4f' }}>{chainData.lastError}</span>
          ) : (
            '-'
          )}
        </Descriptions.Item>
      </Descriptions>

      <Divider>可用干预动作</Divider>
      <Space wrap>
        {isRetryAvailable ? (
          <Tag color="blue">可重试</Tag>
        ) : (
          <Tag color="default">重试（不可用）</Tag>
        )}
        {isReplayAvailable ? (
          <Tag color="cyan">可回放</Tag>
        ) : (
          <Tag color="default">回放（不可用）</Tag>
        )}
        {isPauseAvailable ? (
          <Tag color="orange">可暂停</Tag>
        ) : (
          <Tag color="default">暂停（不可用）</Tag>
        )}
        {isEscalateAvailable ? (
          <Tag color="red">可升级</Tag>
        ) : (
          <Tag color="default">升级（不可用）</Tag>
        )}
      </Space>
    </Drawer>
  );
};

export default ChainDetailDrawer;
