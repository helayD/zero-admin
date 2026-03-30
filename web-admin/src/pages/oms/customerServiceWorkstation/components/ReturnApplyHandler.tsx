import React, { useState } from 'react';
import {
  Button,
  Card,
  Col,
  Input,
  message as antdMessage,
  Modal,
  Row,
  Space,
  Radio,
} from 'antd';
import { ExclamationCircleOutlined, EditOutlined } from '@ant-design/icons';
import type { CustomerServiceOrderItem, CompanyAddressItem } from '../data.d';
import { updateOrderReturn } from '../service';

const { TextArea } = Input;
const { confirm } = Modal;

export interface ReturnApplyHandlerProps {
  order: CustomerServiceOrderItem;
  companyAddresses: CompanyAddressItem[];
  onSuccess: () => void;
}

const ReturnApplyHandler: React.FC<ReturnApplyHandlerProps> = (props) => {
  const { order, companyAddresses, onSuccess } = props;
  const [visible, setVisible] = useState(false);
  const [handleStatus, setHandleStatus] = useState<number>(1);
  const [handleNote, setHandleNote] = useState('');
  const [refundAmount, setRefundAmount] = useState<number>(order.payAmount || 0);
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async () => {
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
        setVisible(false);
        onSuccess();
      } else {
        antdMessage.error(res.message || '操作失败');
      }
    } catch {
      setSubmitting(false);
      antdMessage.error('操作失败，请稍后重试');
    }
  };

  const consequenceMsg =
    handleStatus === 1
      ? `审核通过后，退款金额 ${refundAmount.toFixed(2)} 将记录到 OMS。实际财务退款由 Epic 8 独立财务模块处理。`
      : `审核拒绝后，退货申请将被关闭，用户将收到拒绝通知。`;

  return (
    <>
      <Button
        type="primary"
        icon={<EditOutlined />}
        onClick={() => setVisible(true)}
      >
        处理退货
      </Button>

      <Modal
        title="退货处理"
        open={visible}
        onCancel={() => setVisible(false)}
        footer={[
          <Button key="cancel" onClick={() => setVisible(false)}>
            取消
          </Button>,
          <Button
            key="submit"
            type="primary"
            loading={submitting}
            onClick={handleSubmit}
          >
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
              <div style={{ marginBottom: 8, fontWeight: 500 }}>
                拒绝原因 <span style={{ color: '#f5222d' }}>*</span>
              </div>
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

          <div
            style={{
              background: '#f6ffed',
              padding: 12,
              borderRadius: 4,
              border: '1px solid #b7eb8f',
              fontSize: 12,
              color: '#389e0d',
            }}
          >
            <ExclamationCircleOutlined style={{ marginRight: 8 }} />
            {consequenceMsg}
          </div>
        </Space>
      </Modal>
    </>
  );
};

export default ReturnApplyHandler;
