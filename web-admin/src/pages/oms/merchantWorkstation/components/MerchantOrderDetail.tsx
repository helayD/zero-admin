import React, { useEffect, useState } from 'react';
import {
  Button,
  Card,
  Descriptions,
  Divider,
  Input,
  message,
  Modal,
  Radio,
  Space,
  Tag,
  Timeline,
  message as antdMessage,
} from 'antd';
import { ExclamationCircleOutlined } from '@ant-design/icons';
import type { MerchantOrderItem, MerchantOrderItemData } from '../data';
import type { OrderOperationLogItem, OrderOperationLogListParams } from '../data';
import type { CompanyAddressItem, CompanyAddressListParams } from '../data';
import {
  queryOrderOperationLogList,
  queryCompanyAddressList,
  updateOrderReturn,
  updateMerchantOrderRemark,
} from '../service';
import MerchantDeliveryConfirm from './MerchantDeliveryConfirm';

const { TextArea } = Input;

export interface MerchantOrderDetailProps {
  order: MerchantOrderItem;
  onSuccess: () => void;
  autoOpenDeliveryConfirm?: boolean;
  autoOpenReturnModal?: boolean;
}

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

const MerchantOrderDetail: React.FC<MerchantOrderDetailProps> = (props) => {
  const { order, onSuccess, autoOpenDeliveryConfirm, autoOpenReturnModal } = props;
  const [operationLogs, setOperationLogs] = useState<OrderOperationLogItem[]>([]);
  const [companyAddresses, setCompanyAddresses] = useState<CompanyAddressItem[]>([]);
  const [loadingLogs, setLoadingLogs] = useState(false);
  const [returnModalVisible, setReturnModalVisible] = useState(false);
  const [deliveryModalVisible, setDeliveryModalVisible] = useState(false);
  const [handleStatus, setHandleStatus] = useState<number>(1);
  const [handleNote, setHandleNote] = useState('');
  const [refundAmount, setRefundAmount] = useState<number>(order.payAmount || 0);
  const [submitting, setSubmitting] = useState(false);
  const [merchantNotes, setMerchantNotes] = useState(order.merchantNotes || '');
  const [notesEditing, setNotesEditing] = useState(false);

  useEffect(() => {
    if (autoOpenDeliveryConfirm) {
      setDeliveryModalVisible(true);
    }
  }, [autoOpenDeliveryConfirm]);

  useEffect(() => {
    if (autoOpenReturnModal) {
      setReturnModalVisible(true);
    }
  }, [autoOpenReturnModal]);

  useEffect(() => {
    setLoadingLogs(true);
    queryOrderOperationLogList({
      orderId: order.id,
      current: 1,
      pageSize: 50,
    } as OrderOperationLogListParams).then((res) => {
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
      queryCompanyAddressList({
        current: 1,
        pageSize: 100,
      } as CompanyAddressListParams).then((res) => {
        if (res.code === '000000') {
          setCompanyAddresses(res.data || []);
        }
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
        setReturnModalVisible(false);
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
      return `审核通过后，退款金额 ${refundAmount.toFixed(2)} 将记录到 OMS 系统。实际财务退款由 Epic 8 独立财务模块处理。`;
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
          <Descriptions.Item label="会员昵称">{order.memberNickname || '-'}</Descriptions.Item>
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
          {order.orderItemData.map((item: MerchantOrderItemData) => (
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

      <Card
        title="商户备注"
        style={{ marginBottom: 16 }}
        extra={
          !notesEditing ? (
            <Button size="small" onClick={() => setNotesEditing(true)}>
              编辑备注
            </Button>
          ) : null
        }
      >
        {notesEditing ? (
          <Space direction="vertical" style={{ width: '100%' }} size="middle">
            <TextArea
              rows={3}
              placeholder="请输入商户内部备注"
              value={merchantNotes}
              onChange={(e) => setMerchantNotes(e.target.value)}
            />
            <Space>
              <Button
                type="primary"
                size="small"
                onClick={async () => {
                  try {
                    const res = await updateMerchantOrderRemark({
                      id: order.id,
                      note: merchantNotes,
                    });
                    if (res.code === '000000') {
                      setNotesEditing(false);
                      antdMessage.success('备注已保存');
                    } else {
                      antdMessage.error(res.message || '保存失败');
                    }
                  } catch {
                    antdMessage.error('保存失败，请稍后重试');
                  }
                }}
              >
                保存
              </Button>
              <Button size="small" onClick={() => { setNotesEditing(false); setMerchantNotes(order.merchantNotes || ''); }}>
                取消
              </Button>
            </Space>
          </Space>
        ) : (
          <span style={{ color: merchantNotes ? '#333' : '#999' }}>
            {merchantNotes || '暂无备注'}
          </span>
        )}
      </Card>

      {order.returnStatus !== 0 && (
        <Card
          title="退货信息"
          style={{ marginBottom: 16 }}
          extra={
            <Button
              type="primary"
              onClick={() => setReturnModalVisible(true)}
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

      {order.orderStatus === 2 && !order.deliverySn && (
        <Card style={{ marginBottom: 16 }}>
          <Space>
            <Button type="primary" onClick={() => setDeliveryModalVisible(true)}>
              确认发货
            </Button>
          </Space>
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
        open={returnModalVisible}
        onCancel={() => setReturnModalVisible(false)}
        footer={[
          <Button key="cancel" onClick={() => setReturnModalVisible(false)}>
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

      <MerchantDeliveryConfirm
        order={order}
        visible={deliveryModalVisible}
        onCancel={() => setDeliveryModalVisible(false)}
        onSuccess={() => {
          setDeliveryModalVisible(false);
          onSuccess();
        }}
      />
    </div>
  );
};

export default MerchantOrderDetail;
