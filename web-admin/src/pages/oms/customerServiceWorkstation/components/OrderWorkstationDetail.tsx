import React, { useEffect, useState } from 'react';
import {
  Button,
  Card,
  Col,
  Descriptions,
  Divider,
  message,
  Modal,
  Radio,
  Row,
  Space,
  Steps,
  Tag,
  Timeline,
  Input,
  message as antdMessage,
} from 'antd';
import { ExclamationCircleOutlined } from '@ant-design/icons';
import type { CustomerServiceOrderItem } from '../data.d';
import type { OrderOperationLogItem } from '../data.d';
import type { CompanyAddressItem } from '../data.d';
import type { OrderItemData } from '../data.d';
import {
  queryOrderOperationLogList,
  queryCompanyAddressList,
  updateOrderReturn,
} from '../service';
import type { GovernanceScopeValue } from '@/pages/system/components/governance';
import { toGovernancePayload } from '@/pages/system/components/governance';

const { confirm } = Modal;
const { TextArea } = Input;

export interface OrderWorkstationDetailProps {
  order: CustomerServiceOrderItem;
  scope: GovernanceScopeValue;
  onSuccess: () => void;
  autoOpenReturnModal?: boolean;
}

const statusMeta = (status: number) => {
  switch (status) {
    case 1: return '待支付';
    case 2: return '已支付';
    case 3: return '已发货';
    case 4: return '已完成';
    case 5: return '已取消';
    case 6: return '已退款';
    case 7: return '售后中';
    default: return '未知';
  }
};

const getOrderStatusColor = (status: number) => {
  switch (status) {
    case 1: return 'default';
    case 2: return 'processing';
    case 3: return 'blue';
    case 4: return 'success';
    case 5: return 'warning';
    case 6: return 'purple';
    case 7: return 'error';
    default: return 'default';
  }
};

const getReturnStatusColor = (status: number) => {
  switch (status) {
    case 0: return 'default';
    case 1: return 'warning';
    case 2: return 'processing';
    case 3: return 'error';
    default: return 'default';
  }
};

const OrderWorkstationDetail: React.FC<OrderWorkstationDetailProps> = (props) => {
  const { order, scope, onSuccess, autoOpenReturnModal } = props;
  const [operationLogs, setOperationLogs] = useState<OrderOperationLogItem[]>([]);
  const [companyAddresses, setCompanyAddresses] = useState<CompanyAddressItem[]>([]);
  const [loadingLogs, setLoadingLogs] = useState(false);
  const [loadingAddresses, setLoadingAddresses] = useState(false);
  const [handleModalVisible, setHandleModalVisible] = useState(false);
  const [handleStatus, setHandleStatus] = useState<number>(1); // 1=通过, 2=拒绝
  const [handleNote, setHandleNote] = useState('');
  const [refundAmount, setRefundAmount] = useState<number>(order.payAmount || 0);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (autoOpenReturnModal) {
      setHandleModalVisible(true);
    }
  }, [autoOpenReturnModal]);

  useEffect(() => {
    setLoadingLogs(true);
    queryOrderOperationLogList({
      orderId: order.id,
      current: 1,
      pageSize: 50,
      ...toGovernancePayload(scope),
    }).then((res) => {
      setLoadingLogs(false);
      if (res.code === '000000') {
        setOperationLogs(res.data || []);
      } else {
        message.error(res.message || '查询操作日志失败');
      }
    }).catch(() => {
      setLoadingLogs(false);
    });
  }, [order.id]);

  useEffect(() => {
    if (order.returnStatus !== 0) {
      setLoadingAddresses(true);
      queryCompanyAddressList({
        current: 1,
        pageSize: 100,
        ...toGovernancePayload(scope),
      }).then((res) => {
        setLoadingAddresses(false);
        if (res.code === '000000') {
          setCompanyAddresses(res.data || []);
        }
      }).catch(() => {
        setLoadingAddresses(false);
      });
    }
  }, [order.returnStatus]);

  const handleSubmitReturn = async () => {
    if (handleStatus === 2 && !handleNote.trim()) {
      antdMessage.error('请填写拒绝原因');
      return;
    }
    setSubmitting(true);
    try {
      const res = await updateOrderReturn({
        id: order.id,
        status: handleStatus,
        handleNote: handleNote,
        refundAmount: handleStatus === 1 ? refundAmount : undefined,
      });
      setSubmitting(false);
      if (res.code === '000000') {
        antdMessage.success(handleStatus === 1 ? '审核通过' : '审核拒绝');
        setHandleModalVisible(false);
        onSuccess();
      } else {
        antdMessage.error(res.message || '操作失败');
      }
    } catch {
      setSubmitting(false);
      antdMessage.error('操作失败，请稍后重试');
    }
  };

  const showConsequencePreview = () => {
    if (handleStatus === 1) {
      return `审核通过后，退款金额 ${refundAmount.toFixed(2)} 将记录到系统中。实际财务退款由 Epic 8 独立财务模块处理。`;
    }
    return `审核拒绝后，退货申请将被关闭，用户将收到拒绝通知。`;
  };

  return (
    <div>
      <Card title="基本信息" style={{ marginBottom: 16 }}>
        <Descriptions column={2}>
          <Descriptions.Item label="订单编号">{order.orderNo}</Descriptions.Item>
          <Descriptions.Item label="订单状态">
            <Tag color={getOrderStatusColor(order.orderStatus)}>{order.orderStatusText}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="支付状态">
            <Tag color={order.payStatus === 1 ? 'success' : order.payStatus === 2 ? 'error' : 'warning'}>
              {order.payStatusText}
            </Tag>
          </Descriptions.Item>
          <Descriptions.Item label="售后状态">
            <Tag color={getReturnStatusColor(order.returnStatus)}>{order.returnStatusText}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="下单时间">{order.createTime || '-'}</Descriptions.Item>
          <Descriptions.Item label="支付时间">{order.payTime || '-'}</Descriptions.Item>
          <Descriptions.Item label="发货时间">{order.deliveryTime || '-'}</Descriptions.Item>
          <Descriptions.Item label="收货时间">{order.receiveTime || '-'}</Descriptions.Item>
        </Descriptions>
      </Card>

      <Card title="会员信息" style={{ marginBottom: 16 }}>
        <Descriptions column={2}>
          <Descriptions.Item label="会员ID">{order.memberId || '-'}</Descriptions.Item>
          <Descriptions.Item label="收货人">{order.receiverName || '-'}</Descriptions.Item>
          <Descriptions.Item label="联系电话">{order.receiverPhone || '-'}</Descriptions.Item>
          <Descriptions.Item label="收货地址" span={2}>{order.receiverAddress || '-'}</Descriptions.Item>
        </Descriptions>
      </Card>

      <Card title="金额信息" style={{ marginBottom: 16 }}>
        <Descriptions column={2}>
          <Descriptions.Item label="订单总金额">¥{order.totalAmount?.toFixed(2) ?? '0.00'}</Descriptions.Item>
          <Descriptions.Item label="实付金额">
            <span style={{ color: '#f5222d', fontWeight: 'bold' }}>¥{order.payAmount?.toFixed(2) ?? '0.00'}</span>
          </Descriptions.Item>
        </Descriptions>
      </Card>

      {(order.orderItemData && order.orderItemData.length > 0) && (
        <Card title="商品明细" style={{ marginBottom: 16 }}>
          {order.orderItemData.map((item: OrderItemData) => (
            <div key={item.id} style={{ display: 'flex', alignItems: 'center', padding: '8px 0', borderBottom: '1px solid #f0f0f0' }}>
              {item.skuPic && (
                <img
                  src={item.skuPic}
                  alt={item.skuName}
                  style={{ width: 60, height: 60, objectFit: 'cover', borderRadius: 4, marginRight: 12 }}
                />
              )}
              <div style={{ flex: 1 }}>
                <div style={{ fontWeight: 500, marginBottom: 4 }}>{item.skuName}</div>
                <div style={{ color: '#888', fontSize: 12 }}>
                  {item.specData ? `${item.specData} · ` : ''}
                  数量: {item.skuQuantity} × ¥{item.skuPrice?.toFixed(2) ?? '0.00'}
                </div>
              </div>
              <div style={{ textAlign: 'right' }}>
                <div style={{ fontWeight: 500, color: '#f5222d' }}>¥{item.realAmount?.toFixed(2) ?? '0.00'}</div>
                {item.orderItemStatus === 2 && (
                  <Tag color="warning" style={{ marginTop: 4 }}>退货申请中</Tag>
                )}
                {item.orderItemStatus === 3 && (
                  <Tag color="success" style={{ marginTop: 4 }}>已退货</Tag>
                )}
                {item.orderItemStatus === 4 && (
                  <Tag color="error" style={{ marginTop: 4 }}>已拒绝</Tag>
                )}
              </div>
            </div>
          ))}
        </Card>
      )}

      {(order.deliveryCompany || order.deliverySn) && (
        <Card title="物流信息" style={{ marginBottom: 16 }}>
          <Descriptions column={2}>
            <Descriptions.Item label="物流公司">{order.deliveryCompany || '-'}</Descriptions.Item>
            <Descriptions.Item label="物流单号">{order.deliverySn || '-'}</Descriptions.Item>
          </Descriptions>
        </Card>
      )}

      {order.returnStatus !== 0 && (
        <Card
          title="退货信息"
          style={{ marginBottom: 16 }}
          extra={
            <Button
              type="primary"
              onClick={() => setHandleModalVisible(true)}
            >
              处理退货
            </Button>
          }
        >
          <Descriptions column={2}>
            <Descriptions.Item label="售后状态">
              <Tag color={getReturnStatusColor(order.returnStatus)}>{order.returnStatusText}</Tag>
            </Descriptions.Item>
          </Descriptions>
          {companyAddresses.length > 0 && (
            <>
              <Divider>公司退货地址</Divider>
              {companyAddresses.map((addr) => (
                <Tag key={addr.id} color={addr.defaultStatus === 1 ? 'blue' : 'default'}>
                  {addr.addressName} | {addr.receiverName} | {addr.phone} | {addr.fullAddress}
                  {addr.defaultStatus === 1 ? ' (默认)' : ''}
                </Tag>
              ))}
            </>
          )}
        </Card>
      )}

      <Card title="操作日志" style={{ marginBottom: 16 }}>
        {loadingLogs ? (
          <antdMessage.loading>加载中...</antdMessage.loading>
        ) : operationLogs.length > 0 ? (
          <Timeline
            items={operationLogs.map((log) => ({
              color: 'blue',
              children: (
                <div>
                  <b>{log.operationTypeText}</b>
                  {' '}({log.operatorTypeText})
                  {' '}{log.createTime}
                  {log.operatorNote && <div style={{ color: '#888', fontSize: 12 }}>备注: {log.operatorNote}</div>}
                </div>
              ),
            }))}
          />
        ) : (
          <span style={{ color: '#999' }}>暂无操作记录</span>
        )}
      </Card>

      <Modal
        title="退货处理"
        open={handleModalVisible}
        onCancel={() => {
          setHandleModalVisible(false);
          // 通知父组件重置自动打开状态
          if (autoOpenReturnModal) {
            // 父组件通过 onSuccess 回调和 close drawer 重置 openReturnModal
          }
        }}
        footer={[
          <Button key="cancel" onClick={() => setHandleModalVisible(false)}>
            取消
          </Button>,
          <Button key="submit" type="primary" loading={submitting} onClick={handleSubmitReturn}>
            确认提交
          </Button>,
        ]}
      >
        <Space direction="vertical" style={{ width: '100%' }} size="middle">
          <div>
            <div style={{ marginBottom: 8, fontWeight: 500 }}>审核结果</div>
            <Radio.Group
              value={handleStatus}
              onChange={(e) => setHandleStatus(e.target.value)}
            >
              <Radio.Button value={1}>审核通过</Radio.Button>
              <Radio.Button value={2}>审核拒绝</Radio.Button>
            </Radio.Group>
          </div>

          {handleStatus === 2 && (
            <div>
              <div style={{ marginBottom: 8, fontWeight: 500 }}>拒绝原因 <span style={{ color: '#f5222d' }}>*</span></div>
              <TextArea
                rows={3}
                placeholder="请填写拒绝原因"
                value={handleNote}
                onChange={(e) => setHandleNote(e.target.value)}
              />
            </div>
          )}

          {handleStatus === 1 && (
            <div>
              <div style={{ marginBottom: 8, fontWeight: 500 }}>退款金额（元）</div>
              <Input
                type="number"
                placeholder="请输入退款金额"
                value={refundAmount}
                onChange={(e) => setRefundAmount(parseFloat(e.target.value) || 0)}
                prefix="¥"
                style={{ width: 200 }}
              />
            </div>
          )}

          <div style={{ background: '#f6ffed', padding: 12, borderRadius: 4, border: '1px solid #b7eb8f' }}>
            <ExclamationCircleOutlined style={{ color: '#52c41a', marginRight: 8 }} />
            <span style={{ color: '#389e0d', fontSize: 12 }}>{showConsequencePreview()}</span>
          </div>
        </Space>
      </Modal>
    </div>
  );
};

export default OrderWorkstationDetail;
