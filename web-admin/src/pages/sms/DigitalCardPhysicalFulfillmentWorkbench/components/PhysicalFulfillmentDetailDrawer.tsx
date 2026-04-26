import React from 'react';
import { Button, Descriptions, Drawer, Empty, Space, Tag, Timeline, Typography } from 'antd';
import {
  ExclamationCircleOutlined,
  PrinterOutlined,
  ReloadOutlined,
  SendOutlined,
} from '@ant-design/icons';
import type { DigitalCardPhysicalFulfillmentDetailData } from '../data';
import { getFulfillmentStatusColor, getTimelineColor } from '../helper';

const { Text } = Typography;

interface PhysicalFulfillmentDetailDrawerProps {
  open: boolean;
  loading?: boolean;
  detail?: DigitalCardPhysicalFulfillmentDetailData;
  onClose: () => void;
  onUpdateProduction: () => void;
  onShip: () => void;
  onException: () => void;
  onReissue: () => void;
}

const PhysicalFulfillmentDetailDrawer: React.FC<PhysicalFulfillmentDetailDrawerProps> = ({
  open,
  loading = false,
  detail,
  onClose,
  onUpdateProduction,
  onShip,
  onException,
  onReissue,
}) => {
  const item = detail?.item;

  return (
    <Drawer
      title="实体卡履约详情"
      placement="right"
      width={760}
      onClose={onClose}
      open={open}
      extra={
        item ? (
          <Space wrap>
            <Button icon={<PrinterOutlined />} onClick={onUpdateProduction}>
              制作状态
            </Button>
            <Button type="primary" icon={<SendOutlined />} onClick={onShip}>
              登记物流
            </Button>
            <Button danger icon={<ExclamationCircleOutlined />} onClick={onException}>
              标记异常
            </Button>
            <Button icon={<ReloadOutlined />} onClick={onReissue}>
              发起补发
            </Button>
          </Space>
        ) : null
      }
    >
      {!item ? (
        <Empty description={loading ? '正在加载履约详情...' : '暂无履约详情'} />
      ) : (
        <Space direction="vertical" size={16} style={{ width: '100%' }}>
          <Descriptions bordered size="small" column={2}>
            <Descriptions.Item label="履约单 ID">{item.fulfillmentId}</Descriptions.Item>
            <Descriptions.Item label="履约单号">
              {item.fulfillmentNo ? <Text code>{item.fulfillmentNo}</Text> : '-'}
            </Descriptions.Item>
            <Descriptions.Item label="资产实例 ID">{item.assetInstanceId}</Descriptions.Item>
            <Descriptions.Item label="资产编号">{item.assetNo || '-'}</Descriptions.Item>
            <Descriptions.Item label="会员 ID">{item.memberId}</Descriptions.Item>
            <Descriptions.Item label="活动">
              {item.activityName || `#${item.activityId}`}
            </Descriptions.Item>
            <Descriptions.Item label="模板">
              {item.templateName || `#${item.templateId}`}
            </Descriptions.Item>
            <Descriptions.Item label="履约状态">
              <Tag color={getFulfillmentStatusColor(item.fulfillmentStatus)}>
                {item.fulfillmentStatusText || item.fulfillmentStatus || '-'}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="制作状态">
              {item.productionStatusText || item.productionStatus || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="物流状态">
              {item.shippingStatusText || item.shippingStatus || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="制作批次">{item.productionBatchNo || '-'}</Descriptions.Item>
            <Descriptions.Item label="最近更新时间">{item.updateTime || '-'}</Descriptions.Item>
          </Descriptions>

          <Descriptions bordered size="small" column={2}>
            <Descriptions.Item label="收件人">{detail.receiverName || '-'}</Descriptions.Item>
            <Descriptions.Item label="手机号">{detail.receiverPhone || '-'}</Descriptions.Item>
            <Descriptions.Item label="收货地址" span={2}>
              {detail.addressSummary || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="承运商">{item.carrierName || '-'}</Descriptions.Item>
            <Descriptions.Item label="物流单号">
              {item.trackingNo ? <Text code>{item.trackingNo}</Text> : '-'}
            </Descriptions.Item>
            <Descriptions.Item label="发货时间">{item.shippedAt || '-'}</Descriptions.Item>
            <Descriptions.Item label="签收时间">{item.signedAt || '-'}</Descriptions.Item>
            <Descriptions.Item label="异常码">{item.failureCode || '-'}</Descriptions.Item>
            <Descriptions.Item label="异常原因">{item.failureReason || '-'}</Descriptions.Item>
            <Descriptions.Item label="requestId" span={2}>
              {detail.requestId ? <Text code>{detail.requestId}</Text> : '-'}
            </Descriptions.Item>
            <Descriptions.Item label="traceId" span={2}>
              {detail.traceId ? <Text code>{detail.traceId}</Text> : '-'}
            </Descriptions.Item>
          </Descriptions>

          {detail.timeline?.length ? (
            <Timeline>
              {detail.timeline.map((log) => (
                <Timeline.Item
                  key={`${log.action}-${log.createTime}-${log.reason}`}
                  color={getTimelineColor(log.action)}
                >
                  <Space direction="vertical" size={4}>
                    <Space wrap>
                      <Text strong>{log.actionText || log.action}</Text>
                      <Tag>{log.statusText || '-'}</Tag>
                      <Text type="secondary">{log.createTime || '-'}</Text>
                    </Space>
                    <Text type="secondary">{log.reason || '无附加说明'}</Text>
                  </Space>
                </Timeline.Item>
              ))}
            </Timeline>
          ) : (
            <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无履约时间线" />
          )}
        </Space>
      )}
    </Drawer>
  );
};

export default PhysicalFulfillmentDetailDrawer;
