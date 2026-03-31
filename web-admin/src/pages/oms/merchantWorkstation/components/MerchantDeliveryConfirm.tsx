import React, { useEffect, useState } from 'react';
import {
  Button,
  Form,
  Input,
  message,
  Modal,
  Select,
  Space,
  Alert,
  Spin,
} from 'antd';
import { ExclamationCircleOutlined } from '@ant-design/icons';
import type { MerchantOrderItem, CompanyAddressItem } from '../data';
import { confirmMerchantDelivery, queryCompanyAddressList } from '../service';

const LOGISTICS_COMPANIES = [
  { label: '顺丰速运', value: '顺丰速运' },
  { label: '中通快递', value: '中通快递' },
  { label: '圆通速递', value: '圆通速递' },
  { label: '申通快递', value: '申通快递' },
  { label: '韵达速递', value: '韵达速递' },
  { label: '京东物流', value: '京东物流' },
  { label: '邮政EMS', value: '邮政EMS' },
  { label: '极兔速递', value: '极兔速递' },
  { label: '德邦快递', value: '德邦快递' },
  { label: '其他', value: '其他' },
];

const TRACKING_REGEX = /^[A-Za-z0-9\-]{6,32}$/;

export interface MerchantDeliveryConfirmProps {
  order: MerchantOrderItem;
  visible: boolean;
  onCancel: () => void;
  onSuccess: () => void;
}

const MerchantDeliveryConfirm: React.FC<MerchantDeliveryConfirmProps> = (props) => {
  const { order, visible, onCancel, onSuccess } = props;
  const [form] = Form.useForm();
  const [submitting, setSubmitting] = useState(false);
  const [loadingAddresses, setLoadingAddresses] = useState(false);

  useEffect(() => {
    if (visible) {
      setLoadingAddresses(true);
      queryCompanyAddressList({ current: 1, pageSize: 50 })
        .then((res) => {
          setLoadingAddresses(false);
          if (res.code === '000000' && res.data) {
            const defaultAddr = res.data.find((a: CompanyAddressItem) => a.sendStatus === 1);
            if (defaultAddr) {
              form.setFieldValue('deliveryCompany', defaultAddr.addressName);
            }
          }
        })
        .catch(() => {
          setLoadingAddresses(false);
        });
    } else {
      form.resetFields();
    }
  }, [visible]);

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      setSubmitting(true);
      try {
        const res = await confirmMerchantDelivery({
          orderId: order.id,
          deliveryCompany: values.deliveryCompany,
          deliverySn: values.deliverySn,
        });
        setSubmitting(false);
        if (res.code === '000000') {
          message.success('发货成功');
          form.resetFields();
          onSuccess();
        } else {
          message.error(res.message || '发货失败');
        }
      } catch {
        setSubmitting(false);
        message.error('发货失败，请稍后重试');
      }
    } catch {
      // validation failed
    }
  };

  return (
    <Modal
      title="确认发货"
      open={visible}
      onCancel={onCancel}
      footer={[
        <Button key="cancel" onClick={onCancel} disabled={submitting}>
          取消
        </Button>,
        <Button
          key="submit"
          type="primary"
          loading={submitting}
          onClick={handleSubmit}
        >
          确认发货
        </Button>,
      ]}
      destroyOnClose
    >
      <Space direction="vertical" style={{ width: '100%' }} size="middle">
        <Alert
          message="发货前请确认"
          description={`确认发货后，订单状态将更新为「已发货」，买家可查看物流信息。订单编号：${order.orderNo}`}
          type="info"
          showIcon
        />

        {order.receiverAddress ? null : (
          <Alert
            message="收货地址不完整"
            description="该订单缺少收货地址，请先完善收货信息后再发货。"
            type="warning"
            showIcon
          />
        )}

        <Spin spinning={loadingAddresses}>
          <Form form={form} layout="vertical" requiredMark="optional">
            <Form.Item
              name="deliveryCompany"
              label="物流公司"
              rules={[{ required: true, message: '请选择物流公司' }]}
            >
              <Select
                placeholder="请选择物流公司"
                options={LOGISTICS_COMPANIES}
                showSearch
                optionFilterProp="label"
              />
            </Form.Item>

            <Form.Item
              name="deliverySn"
              label="运单号"
              rules={[
                { required: true, message: '请输入运单号' },
                {
                  pattern: TRACKING_REGEX,
                  message: '运单号格式不正确（6-32位字母或数字）',
                },
              ]}
            >
              <Input placeholder="请输入物流单号" maxLength={32} />
            </Form.Item>
          </Form>
        </Spin>

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
          确认发货后，订单状态将更新为「已发货」，买家可查看物流信息。本次操作将记录到操作日志中。
        </div>
      </Space>
    </Modal>
  );
};

export default MerchantDeliveryConfirm;
