import React, { useEffect, useState } from 'react';
import {
  Form,
  Input,
  InputNumber,
  Switch,
  Select,
  Modal,
  message,
  Row,
  Col,
  Tag,
} from 'antd';
import type { ProductFulfillmentRuleListItem } from '../data.d';
import { RefundPolicyOptions } from '../data.d';
import {
  addProductFulfillmentRule,
  updateProductFulfillmentRule,
  queryProductFulfillmentRuleDetail,
} from '../service';
import type { GovernanceScopeValue } from '@/pages/system/components/governance';
import { toGovernancePayload } from '@/pages/system/components/governance';

interface RuleFormProps {
  visible: boolean;
  onCancel: () => void;
  onSuccess: () => void;
  record?: ProductFulfillmentRuleListItem;
  scope: GovernanceScopeValue;
}

const cardTemplateOptions = [
  { label: 'SSR 银河兔', value: 920001 },
  { label: 'SR 星云狐', value: 920002 },
  { label: '新手欢迎徽章', value: 920003 },
];

const RuleForm: React.FC<RuleFormProps> = ({
  visible,
  onCancel,
  onSuccess,
  record,
  scope,
}) => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [transferable, setTransferable] = useState(false);
  const isEdit = !!record;

  useEffect(() => {
    if (visible) {
      if (isEdit && record) {
        setTransferable(record.transferable === 1);
        form.setFieldsValue({
          ruleName: record.ruleName,
          cardTemplateId: record.cardTemplateId,
          expireDays: record.expireDays,
          transferable: record.transferable === 1,
          transferLimit: record.transferLimit,
          claimCondition: record.claimCondition,
          redemptionCondition: record.redemptionCondition,
          refundPolicy: record.refundPolicy,
        });
      } else {
        setTransferable(false);
        form.resetFields();
        form.setFieldsValue({
          transferable: false,
          expireDays: 365,
          transferLimit: 0,
          refundPolicy: 'freeze_card',
        });
      }
    }
  }, [visible, record, isEdit, form]);

  // 编辑时加载最新详情
  useEffect(() => {
    if (visible && isEdit && record) {
      setLoading(true);
      queryProductFulfillmentRuleDetail(record.id, toGovernancePayload(scope))
        .then((res) => {
          if (res.data) {
            const d = res.data;
            setTransferable(d.transferable === 1);
            form.setFieldsValue({
              ruleName: d.ruleName,
              cardTemplateId: d.cardTemplateId,
              expireDays: d.expireDays,
              transferable: d.transferable === 1,
              transferLimit: d.transferLimit,
              claimCondition: d.claimCondition,
              redemptionCondition: d.redemptionCondition,
              refundPolicy: d.refundPolicy,
            });
          }
        })
        .catch(() => {
          // 使用列表传入的数据兜底
        })
        .finally(() => setLoading(false));
    }
  }, [visible, isEdit, record, scope, form]);

  const handleSubmit = async () => {
    const values = await form.validateFields();
    setLoading(true);
    try {
      const payload = {
        ruleName: values.ruleName,
        cardTemplateId: values.cardTemplateId,
        expireDays: values.expireDays,
        transferable: values.transferable ? 1 : 0,
        transferLimit: values.transferable ? values.transferLimit || 0 : 0,
        claimCondition: values.claimCondition || '',
        redemptionCondition: values.redemptionCondition || '',
        refundPolicy: values.refundPolicy,
        ...toGovernancePayload(scope),
      };

      if (isEdit && record) {
        await updateProductFulfillmentRule({
          id: record.id,
          ...payload,
        });
        message.success('更新成功');
      } else {
        await addProductFulfillmentRule(payload);
        message.success('新增成功');
      }
      onSuccess();
    } catch (error) {
      message.error(isEdit ? '更新失败' : '新增失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal
      title={isEdit ? '编辑发卡规则' : '新增发卡规则'}
      visible={visible}
      onCancel={onCancel}
      onOk={handleSubmit}
      confirmLoading={loading}
      width={720}
      destroyOnClose
    >
      <Form
        form={form}
        layout="vertical"
        initialValues={{
          transferable: false,
          expireDays: 365,
          transferLimit: 0,
          refundPolicy: 'freeze_card',
        }}
      >
        <Row gutter={16}>
          <Col span={12}>
            <Form.Item
              label="规则名称"
              name="ruleName"
              rules={[{ required: true, message: '请输入规则名称' }]}
            >
              <Input placeholder="请输入规则名称" maxLength={50} />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item
              label="关联卡片模板"
              name="cardTemplateId"
              rules={[{ required: true, message: '请选择卡片模板' }]}
            >
              <Select
                placeholder="请选择卡片模板"
                options={cardTemplateOptions}
                showSearch
                optionFilterProp="label"
              />
            </Form.Item>
          </Col>
        </Row>

        <Row gutter={16}>
          <Col span={12}>
            <Form.Item
              label="有效期（天）"
              name="expireDays"
              rules={[{ required: true, message: '请输入有效期' }]}
            >
              <InputNumber
                style={{ width: '100%' }}
                min={1}
                max={9999}
                placeholder="请输入有效期天数"
              />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item
              label="退款处置策略"
              name="refundPolicy"
              rules={[{ required: true, message: '请选择退款处置策略' }]}
            >
              <Select placeholder="请选择退款处置策略" options={RefundPolicyOptions} />
            </Form.Item>
          </Col>
        </Row>

        <Row gutter={16} align="middle">
          <Col span={8}>
            <Form.Item
              label="是否可转赠"
              name="transferable"
              valuePropName="checked"
            >
              <Switch
                onChange={(checked) => {
                  setTransferable(checked);
                  if (!checked) {
                    form.setFieldsValue({ transferLimit: 0 });
                  }
                }}
              />
            </Form.Item>
          </Col>
          <Col span={16}>
            {transferable && (
              <Form.Item
                label="最大转赠次数"
                name="transferLimit"
                rules={[{ required: true, message: '请输入最大转赠次数' }]}
              >
                <InputNumber
                  style={{ width: '100%' }}
                  min={1}
                  max={99}
                  placeholder="请输入最大转赠次数"
                />
              </Form.Item>
            )}
          </Col>
        </Row>

        <Form.Item label="领取条件（JSON）" name="claimCondition">
          <Input.TextArea
            rows={3}
            placeholder='{"minLevel": 1, "maxClaimsPerUser": 1}'
          />
        </Form.Item>

        <Form.Item label="提货条件" name="redemptionCondition">
          <Input.TextArea
            rows={2}
            placeholder="请输入提货条件说明"
          />
        </Form.Item>

        {isEdit && record && (
          <Form.Item label="当前状态">
            <Tag color={record.ruleStatus === 1 ? 'green' : 'red'}>
              {record.ruleStatus === 1 ? '启用' : '禁用'}
            </Tag>
          </Form.Item>
        )}
      </Form>
    </Modal>
  );
};

export default RuleForm;
