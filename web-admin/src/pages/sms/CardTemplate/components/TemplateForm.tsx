import React, { useEffect, useState } from 'react';
import { Form, Input, InputNumber, Modal, Row, Col, Select, message, Tag } from 'antd';
import type { CardTemplateListItem } from '../data.d';
import { ContentAuditStatusOptions, RarityOptions } from '../data.d';
import {
  addCardTemplate,
  updateCardTemplate,
  queryCardTemplateDetail,
} from '../service';
import UploadFileComponents from '@/components/common/UploadFileComponents';
import type { GovernanceScopeValue } from '@/pages/system/components/governance';
import { toGovernancePayload } from '@/pages/system/components/governance';

interface TemplateFormProps {
  visible: boolean;
  onCancel: () => void;
  onSuccess: () => void;
  record?: CardTemplateListItem;
  scope: GovernanceScopeValue;
}

const displayStatusOptions = [
  { value: 1, label: '上架' },
  { value: 0, label: '下架' },
];

const TemplateForm: React.FC<TemplateFormProps> = ({
  visible,
  onCancel,
  onSuccess,
  record,
  scope,
}) => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const isEdit = !!record;

  useEffect(() => {
    if (!visible) return;
    if (isEdit && record) {
      form.setFieldsValue({
        templateCode: record.templateCode,
        templateName: record.templateName,
        cardFaceImage: record.cardFaceImage,
        copyrightOwner: record.copyrightOwner,
        copyrightProofSummary: record.copyrightProofSummary,
        rarity: record.rarity,
        issueLimit: record.issueLimit,
        displayCopy: record.displayCopy,
        circulationLimitSummary: record.circulationLimitSummary,
        displayStatus: record.displayStatus,
        contentAuditStatus: record.contentAuditStatus,
        providerCode: record.providerCode,
        credentialRef: record.credentialRef,
      });
    } else {
      form.resetFields();
      form.setFieldsValue({
        displayStatus: 1,
        contentAuditStatus: 0,
        issueLimit: 0,
        rarity: 'R',
      });
    }
    // Story 10.10 修复 L3: 用 record?.id 而非 record 引用
  }, [visible, record?.id, isEdit, form]);

  // 编辑场景拉一遍最新数据，覆盖列表数据
  useEffect(() => {
    if (visible && isEdit && record) {
      setLoading(true);
      queryCardTemplateDetail(record.id, toGovernancePayload(scope))
        .then((res) => {
          if (res.data) {
            form.setFieldsValue(res.data);
          }
        })
        .catch(() => {
          // 用列表数据兜底
        })
        .finally(() => setLoading(false));
    }
    // Story 10.10 修复 L3: 用 record?.id 而非 record 引用
  }, [visible, isEdit, record?.id, scope, form]);

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      setLoading(true);
      const payload = {
        templateCode: values.templateCode,
        templateName: values.templateName,
        cardFaceImage: values.cardFaceImage || '',
        copyrightOwner: values.copyrightOwner || '',
        copyrightProofSummary: values.copyrightProofSummary || '',
        rarity: values.rarity || '',
        issueLimit: Number(values.issueLimit || 0),
        displayCopy: values.displayCopy || '',
        circulationLimitSummary: values.circulationLimitSummary || '',
        displayStatus: Number(values.displayStatus ?? 1),
        contentAuditStatus: Number(values.contentAuditStatus ?? 0),
        providerCode: values.providerCode || '',
        credentialRef: values.credentialRef || '',
        ...toGovernancePayload(scope),
      };
      if (isEdit && record) {
        await updateCardTemplate({ id: record.id, ...payload });
        message.success('更新成功');
      } else {
        await addCardTemplate(payload);
        message.success('新增成功');
      }
      onSuccess();
    } catch (error: any) {
      if (error?.errorFields) return; // 表单校验错误已显示
      message.error(isEdit ? '更新失败' : '新增失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal
      title={isEdit ? '编辑卡片模板' : '新增卡片模板'}
      visible={visible}
      onCancel={onCancel}
      onOk={handleSubmit}
      confirmLoading={loading}
      width={780}
      destroyOnClose
    >
      <Form form={form} layout="vertical">
        <Row gutter={16}>
          <Col span={12}>
            <Form.Item
              label="模板编码"
              name="templateCode"
              rules={[
                { required: true, message: '请输入模板编码' },
                { max: 64, message: '不超过 64 字符' },
              ]}
              tooltip="作用域内唯一，创建后不可修改"
            >
              <Input placeholder="例如 TPL-SSR-RABBIT" disabled={isEdit} />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item
              label="模板名称"
              name="templateName"
              rules={[{ required: true, message: '请输入模板名称' }]}
            >
              <Input placeholder="例如 SSR 银河兔" maxLength={100} />
            </Form.Item>
          </Col>
        </Row>

        <Row gutter={16}>
          <Col span={12}>
            <Form.Item label="稀有度" name="rarity">
              <Select options={RarityOptions} placeholder="请选择稀有度" allowClear />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item
              label="发行上限"
              name="issueLimit"
              tooltip="0 表示不限"
            >
              <InputNumber style={{ width: '100%' }} min={0} max={9999999} />
            </Form.Item>
          </Col>
        </Row>

        <Form.Item label="卡面资源" name="cardFaceImage">
          <UploadFileComponents count={1} />
        </Form.Item>

        <Row gutter={16}>
          <Col span={12}>
            <Form.Item label="版权归属" name="copyrightOwner">
              <Input placeholder="例如 九克城" maxLength={128} />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="版权凭证摘要" name="copyrightProofSummary">
              <Input placeholder="版权凭证号或链上 hash 摘要" maxLength={255} />
            </Form.Item>
          </Col>
        </Row>

        <Form.Item label="展示文案" name="displayCopy">
          <Input.TextArea rows={2} maxLength={500} placeholder="C 端卡详情展示文案" />
        </Form.Item>

        <Form.Item label="默认流转限制" name="circulationLimitSummary">
          <Input.TextArea rows={2} maxLength={500} placeholder="例如：每用户最多领取 1 张，30 天内不可转赠" />
        </Form.Item>

        <Row gutter={16}>
          <Col span={12}>
            <Form.Item label="展示状态" name="displayStatus">
              <Select options={displayStatusOptions} />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="内容审核状态" name="contentAuditStatus">
              <Select options={ContentAuditStatusOptions} />
            </Form.Item>
          </Col>
        </Row>

        <Row gutter={16}>
          <Col span={12}>
            <Form.Item label="接入方编码" name="providerCode" tooltip="链上接入方/服务商编码">
              <Input placeholder="例如 antchain / fisco" maxLength={64} />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="凭证引用" name="credentialRef" tooltip="实名/版权凭证引用">
              <Input maxLength={128} />
            </Form.Item>
          </Col>
        </Row>

        {isEdit && record && (
          <Form.Item label="当前状态">
            <Tag color={record.status === 1 ? 'green' : 'red'}>
              {record.status === 1 ? '启用' : '禁用'}
            </Tag>
            <Tag color="blue">引用规则数：{record.refRuleCount || 0}</Tag>
          </Form.Item>
        )}
      </Form>
    </Modal>
  );
};

export default TemplateForm;
