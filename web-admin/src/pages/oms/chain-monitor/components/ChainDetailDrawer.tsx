import React from 'react';
import { Descriptions, Drawer } from 'antd';
import type { ChainMonitorItem } from '../data.d';

interface ChainDetailDrawerProps {
  visible: boolean;
  chainData?: ChainMonitorItem;
  onClose: () => void;
}

const ChainDetailDrawer: React.FC<ChainDetailDrawerProps> = ({
  visible,
  chainData,
  onClose,
}) => {
  if (!chainData) return null;

  return (
    <Drawer
      title="链路详情"
      placement="right"
      width={600}
      onClose={onClose}
      open={visible}
    >
      <Descriptions column={2} bordered size="small">
        <Descriptions.Item label="链路追踪ID" span={2}>
          <span style={{ fontFamily: 'monospace' }}>{chainData.traceId}</span>
        </Descriptions.Item>

        <Descriptions.Item label="平台ID">{chainData.platformId}</Descriptions.Item>
        <Descriptions.Item label="租户ID">{chainData.tenantId}</Descriptions.Item>
        <Descriptions.Item label="商户ID">{chainData.merchantId}</Descriptions.Item>

        <Descriptions.Item label="链路类型">{chainData.chainTypeText}</Descriptions.Item>
        <Descriptions.Item label="业务对象类型">
          {chainData.entityType === 1 ? '订单' : chainData.entityType === 2 ? '商品' : '-'}
        </Descriptions.Item>

        <Descriptions.Item label="业务编号">{chainData.entityNo}</Descriptions.Item>
        <Descriptions.Item label="业务ID">{chainData.entityId}</Descriptions.Item>

        <Descriptions.Item label="当前阶段">{chainData.stageText}</Descriptions.Item>
        <Descriptions.Item label="处理结果">{chainData.resultText}</Descriptions.Item>

        <Descriptions.Item label="重试次数">{chainData.retryCount}</Descriptions.Item>
        <Descriptions.Item label="触发操作人ID">{chainData.actorId || '-'}</Descriptions.Item>

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
    </Drawer>
  );
};

export default ChainDetailDrawer;
