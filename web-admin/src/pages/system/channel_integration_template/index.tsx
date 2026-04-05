import {
  CheckCircleOutlined,
  EditOutlined,
  EyeOutlined,
  InboxOutlined,
  PauseCircleOutlined,
  PlusOutlined,
} from '@ant-design/icons';
import {
  Alert,
  Button,
  Descriptions,
  Divider,
  Drawer,
  Form,
  Input,
  InputNumber,
  message,
  Modal,
  Select,
  Space,
  Tag,
} from 'antd';
import React, { useMemo, useRef, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import type { ActionType, ProColumns } from '@ant-design/pro-table';
import ProTable from '@ant-design/pro-table';
import {
  formatSystemChannelLabel,
  normalizeSystemChannel,
  SYSTEM_CHANNEL_OPTIONS,
  SYSTEM_CHANNEL_VALUE_ENUM,
} from '../shared/channel';
import type {
  ChannelIntegrationTemplateItem,
  ChannelIntegrationTemplateSubmitPayload,
  ChannelTemplateScopeType,
  ChannelTemplateType,
} from './data.d';
import {
  createChannelIntegrationTemplate,
  queryChannelIntegrationTemplateDetail,
  queryChannelIntegrationTemplateList,
  updateChannelIntegrationTemplate,
  updateChannelIntegrationTemplateStatus,
} from './service';

const { TextArea } = Input;
const { confirm } = Modal;

const templateTypeValueEnum = {
  channel: { text: '渠道模板' },
  integration: { text: '三方集成模板' },
};

const scopeTypeValueEnum = {
  platform: { text: '平台级' },
  tenant: { text: '租户级' },
  merchant: { text: '商户级' },
};

const statusValueEnum = {
  draft: { text: '草稿' },
  enabled: { text: '已启用' },
  disabled: { text: '已停用' },
  archived: { text: '已归档' },
};

const statusColorMap: Record<ChannelIntegrationTemplateStatus, string> = {
  draft: 'processing',
  enabled: 'success',
  disabled: 'default',
  archived: 'warning',
};

const integrationTargetOptions = [
  { label: '物流轨迹', value: 'logistics_tracking' },
  { label: '短信服务商', value: 'sms_provider' },
  { label: '会员消息', value: 'member_message' },
];

type ChannelIntegrationTemplateStatus = ChannelIntegrationTemplateItem['status'];

const integrationTargetValueEnum: Record<string, { text: string }> = {
  logistics_tracking: { text: '物流轨迹' },
  sms_provider: { text: '短信服务商' },
  member_message: { text: '会员消息' },
};

const readErrorMessage = (error: any, fallback: string) => {
  return error?.data?.message || error?.message || fallback;
};

const formatJSONText = (value?: any) => {
  if (value === null || value === undefined || value === '') {
    return '{}';
  }
  try {
    return JSON.stringify(value, null, 2);
  } catch (error) {
    return '{}';
  }
};

const parseJSONInput = (value?: string) => {
  const trimmed = (value || '').trim();
  if (!trimmed) {
    return {};
  }
  return JSON.parse(trimmed);
};

const formatTargetLabel = (record: ChannelIntegrationTemplateItem) => {
  if (record.templateType === 'channel') {
    return formatSystemChannelLabel(record.targetCode);
  }
  return integrationTargetValueEnum[record.targetCode]?.text || record.targetCode || '-';
};

const formatImpactSummary = (record: ChannelIntegrationTemplateItem) => {
  const tenantCount = record.bindingTenantCount || 0;
  const merchantCount = record.bindingMerchantCount || 0;
  const labels = record.bindingSampleLabels?.filter(Boolean) || [];
  const parts = [`租户 ${tenantCount}`, `商户 ${merchantCount}`];
  if (labels.length > 0) {
    parts.push(`样本 ${labels.join('、')}`);
  }
  return parts.join(' / ');
};

const TemplatePage: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const [form] = Form.useForm();
  const [modalVisible, setModalVisible] = useState(false);
  const [detailVisible, setDetailVisible] = useState(false);
  const [modalLoading, setModalLoading] = useState(false);
  const [detailLoading, setDetailLoading] = useState(false);
  const [isEditMode, setIsEditMode] = useState(false);
  const [selectedRows, setSelectedRows] = useState<ChannelIntegrationTemplateItem[]>([]);
  const [currentRow, setCurrentRow] = useState<ChannelIntegrationTemplateItem>();
  const [editingRow, setEditingRow] = useState<ChannelIntegrationTemplateItem>();
  const [templateType, setTemplateType] = useState<ChannelTemplateType>('channel');

  const targetOptions = useMemo(() => {
    if (templateType === 'channel') {
      return SYSTEM_CHANNEL_OPTIONS;
    }
    return integrationTargetOptions;
  }, [templateType]);

  const reloadTable = () => {
    actionRef.current?.reload?.();
    setSelectedRows([]);
  };

  const openCreateModal = () => {
    setIsEditMode(false);
    setEditingRow(undefined);
    setTemplateType('channel');
    form.setFieldsValue({
      templateType: 'channel',
      scopeType: 'platform',
      platformId: 1,
      status: 'draft',
      metadataConfigText: '{}',
      secretRefConfigText: '{}',
      intentContractConfigText: '{}',
      impactScopeConfigText: '{}',
    });
    setModalVisible(true);
  };

  const openEditModal = (record: ChannelIntegrationTemplateItem) => {
    setIsEditMode(true);
    setEditingRow(record);
    setTemplateType(record.templateType);
    form.setFieldsValue({
      ...record,
      targetCode:
        record.templateType === 'channel' ? normalizeSystemChannel(record.targetCode) : record.targetCode,
      metadataConfigText: formatJSONText(record.metadataConfig),
      secretRefConfigText: formatJSONText(record.secretRefConfig),
      intentContractConfigText: formatJSONText(record.intentContractConfig),
      impactScopeConfigText: formatJSONText(record.impactScopeConfig),
    });
    setModalVisible(true);
  };

  const openDetail = async (record: ChannelIntegrationTemplateItem) => {
    setDetailVisible(true);
    setDetailLoading(true);
    try {
      const result = await queryChannelIntegrationTemplateDetail(record.id);
      setCurrentRow(result.data);
    } catch (error) {
      message.error(readErrorMessage(error, '加载模板详情失败，请稍后重试'));
      setDetailVisible(false);
    } finally {
      setDetailLoading(false);
    }
  };

  const handleStatusAction = async (status: ChannelIntegrationTemplateStatus, rows: ChannelIntegrationTemplateItem[]) => {
    const hide = message.loading('正在更新模板状态');
    try {
      await updateChannelIntegrationTemplateStatus({ ids: rows.map((item) => item.id), status });
      hide();
      message.success('状态更新成功');
      reloadTable();
      if (currentRow && rows.some((item) => item.id === currentRow.id)) {
        openDetail(currentRow);
      }
      return true;
    } catch (error) {
      hide();
      message.error(readErrorMessage(error, '状态更新失败，请核对模板当前生命周期'));
      return false;
    }
  };

  const showStatusConfirm = (status: ChannelIntegrationTemplateStatus, rows: ChannelIntegrationTemplateItem[]) => {
    const titleMap: Record<ChannelIntegrationTemplateStatus, string> = {
      draft: '回退草稿',
      enabled: '启用',
      disabled: '停用',
      archived: '归档',
    };
    confirm({
      title: `确认${titleMap[status]}${rows.length > 1 ? `这 ${rows.length} 个` : '该'}模板吗？`,
      icon:
        status === 'enabled' ? <CheckCircleOutlined /> : status === 'disabled' ? <PauseCircleOutlined /> : <InboxOutlined />,
      content:
        status === 'archived'
          ? '归档后模板将退出后续治理流转，但审计轨迹会保留。'
          : status === 'disabled'
          ? '停用后不会再继续分发给新的租户/商户绑定。'
          : '启用后模板可以作为后续治理配置的候选项。',
      onOk: async () => handleStatusAction(status, rows),
    });
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      const payload: ChannelIntegrationTemplateSubmitPayload = {
        id: editingRow?.id,
        templateCode: values.templateCode,
        templateName: values.templateName,
        templateType: values.templateType,
        targetCode:
          values.templateType === 'channel' ? normalizeSystemChannel(values.targetCode) : values.targetCode,
        scopeType: values.scopeType,
        platformId: values.platformId,
        tenantId: values.tenantId,
        merchantId: values.merchantId,
        status: values.status,
        metadataConfig: parseJSONInput(values.metadataConfigText),
        secretRefConfig: parseJSONInput(values.secretRefConfigText),
        intentContractConfig: parseJSONInput(values.intentContractConfigText),
        impactScopeConfig: parseJSONInput(values.impactScopeConfigText),
        remark: values.remark,
      };
      setModalLoading(true);
      if (isEditMode && editingRow?.id) {
        await updateChannelIntegrationTemplate(payload);
        message.success('模板更新成功');
      } else {
        await createChannelIntegrationTemplate(payload);
        message.success('模板创建成功');
      }
      setModalVisible(false);
      setEditingRow(undefined);
      form.resetFields();
      reloadTable();
    } catch (error: any) {
      if (error?.errorFields) {
        return;
      }
      message.error(readErrorMessage(error, '保存模板失败，请检查 JSON 配置格式'));
    } finally {
      setModalLoading(false);
    }
  };

  const columns: ProColumns<ChannelIntegrationTemplateItem>[] = [
    {
      title: '模板名称',
      dataIndex: 'templateName',
      render: (_, record) => <a onClick={() => openDetail(record)}>{record.templateName}</a>,
    },
    {
      title: '模板编码',
      dataIndex: 'templateCode',
    },
    {
      title: '模板类型',
      dataIndex: 'templateType',
      valueType: 'select',
      valueEnum: templateTypeValueEnum,
      render: (_, record) => <Tag>{templateTypeValueEnum[record.templateType]?.text || record.templateType}</Tag>,
    },
    {
      title: '目标编码',
      dataIndex: 'targetCode',
      valueType: 'select',
      valueEnum: {
        ...SYSTEM_CHANNEL_VALUE_ENUM,
        ...integrationTargetValueEnum,
      },
      render: (_, record) => formatTargetLabel(record),
    },
    {
      title: '作用域',
      dataIndex: 'scopeType',
      valueType: 'select',
      valueEnum: scopeTypeValueEnum,
    },
    {
      title: '状态',
      dataIndex: 'status',
      valueType: 'select',
      valueEnum: statusValueEnum,
      render: (_, record) => <Tag color={statusColorMap[record.status]}>{statusValueEnum[record.status]?.text}</Tag>,
    },
    {
      title: '租户 ID',
      dataIndex: 'tenantId',
      valueType: 'digit',
      hideInTable: true,
    },
    {
      title: '商户 ID',
      dataIndex: 'merchantId',
      valueType: 'digit',
      hideInTable: true,
    },
    {
      title: '影响范围',
      dataIndex: 'bindingTenantCount',
      hideInSearch: true,
      render: (_, record) => formatImpactSummary(record),
    },
    {
      title: '最近更新',
      dataIndex: 'updateTime',
      hideInSearch: true,
      render: (_, record) => record.updateTime || record.createTime || '-',
    },
    {
      title: '操作',
      valueType: 'option',
      width: 260,
      render: (_, record) => {
        const actions: React.ReactNode[] = [
          <a key="detail" onClick={() => openDetail(record)}>
            <EyeOutlined /> 详情
          </a>,
          <a key="edit" onClick={() => openEditModal(record)}>
            <EditOutlined /> 编辑
          </a>,
        ];
        if (record.status !== 'enabled' && record.status !== 'archived') {
          actions.push(
            <a key="enable" onClick={() => showStatusConfirm('enabled', [record])}>
              <CheckCircleOutlined /> 启用
            </a>,
          );
        }
        if (record.status !== 'disabled' && record.status !== 'archived') {
          actions.push(
            <a key="disable" onClick={() => showStatusConfirm('disabled', [record])}>
              <PauseCircleOutlined /> 停用
            </a>,
          );
        }
        if (record.status !== 'archived') {
          actions.push(
            <a key="archive" onClick={() => showStatusConfirm('archived', [record])}>
              <InboxOutlined /> 归档
            </a>,
          );
        }
        return <Space split={<Divider type="vertical" />}>{actions}</Space>;
      },
    },
  ];

  return (
    <PageContainer>
      <Alert
        showIcon
        type="info"
        style={{ marginBottom: 16 }}
        message="模板治理工作台"
        description="这里统一治理渠道模板与三方集成模板。当前先提供列表、详情、创建、编辑和生命周期流转骨架，后续可继续补充绑定关系、差异比对和审计视图。"
      />

      <ProTable<ChannelIntegrationTemplateItem>
        headerTitle="渠道与集成模板治理"
        actionRef={actionRef}
        rowKey="id"
        search={{ labelWidth: 120 }}
        request={queryChannelIntegrationTemplateList}
        columns={columns}
        pagination={{ pageSize: 20 }}
        rowSelection={{
          onChange: (_, rows) => setSelectedRows(rows),
        }}
        toolBarRender={() => [
          <Button key="create" type="primary" onClick={openCreateModal}>
            <PlusOutlined /> 新建模板
          </Button>,
        ]}
        tableAlertOptionRender={() => {
          if (selectedRows.length === 0) {
            return false;
          }
          return (
            <Space size={16}>
              <a onClick={() => showStatusConfirm('enabled', selectedRows)}>批量启用</a>
              <a onClick={() => showStatusConfirm('disabled', selectedRows)}>批量停用</a>
              <a onClick={() => showStatusConfirm('archived', selectedRows)}>批量归档</a>
            </Space>
          );
        }}
      />

      <Modal
        destroyOnClose
        maskClosable={false}
        confirmLoading={modalLoading}
        open={modalVisible}
        title={isEditMode ? '编辑模板' : '新建模板'}
        width={760}
        onCancel={() => {
          setModalVisible(false);
          setEditingRow(undefined);
          form.resetFields();
        }}
        onOk={handleSubmit}
      >
        <Form
          form={form}
          layout="vertical"
          onValuesChange={(changedValues) => {
            if (changedValues.templateType) {
              setTemplateType(changedValues.templateType as ChannelTemplateType);
              form.setFieldsValue({ targetCode: undefined });
            }
          }}
        >
          <Form.Item label="模板名称" name="templateName" rules={[{ required: true, message: '请输入模板名称' }]}>
            <Input maxLength={64} />
          </Form.Item>
          <Form.Item label="模板编码" name="templateCode" rules={[{ required: true, message: '请输入模板编码' }]}>
            <Input maxLength={64} />
          </Form.Item>
          <Space style={{ display: 'flex' }} align="start">
            <Form.Item label="模板类型" name="templateType" rules={[{ required: true, message: '请选择模板类型' }]}>
              <Select style={{ width: 160 }} options={[
                { label: '渠道模板', value: 'channel' },
                { label: '三方集成模板', value: 'integration' },
              ]} />
            </Form.Item>
            <Form.Item label="目标编码" name="targetCode" rules={[{ required: true, message: '请选择目标编码' }]}>
              <Select style={{ width: 220 }} showSearch options={targetOptions} />
            </Form.Item>
            <Form.Item label="作用域" name="scopeType" rules={[{ required: true, message: '请选择作用域' }]}>
              <Select
                style={{ width: 160 }}
                options={[
                  { label: '平台级', value: 'platform' },
                  { label: '租户级', value: 'tenant' },
                  { label: '商户级', value: 'merchant' },
                ]}
              />
            </Form.Item>
            <Form.Item label="状态" name="status" rules={[{ required: true, message: '请选择状态' }]}>
              <Select
                style={{ width: 140 }}
                options={[
                  { label: '草稿', value: 'draft' },
                  { label: '已启用', value: 'enabled' },
                  { label: '已停用', value: 'disabled' },
                  { label: '已归档', value: 'archived' },
                ]}
              />
            </Form.Item>
          </Space>

          <Space style={{ display: 'flex' }} align="start">
            <Form.Item label="平台 ID" name="platformId">
              <InputNumber min={1} style={{ width: 160 }} />
            </Form.Item>
            <Form.Item noStyle shouldUpdate>
              {({ getFieldValue }) => {
                const scopeType = getFieldValue('scopeType') as ChannelTemplateScopeType;
                if (scopeType === 'platform') {
                  return null;
                }
                return (
                  <Form.Item
                    label="租户 ID"
                    name="tenantId"
                    rules={[{ required: true, message: '请输入租户 ID' }]}
                  >
                    <InputNumber min={1} style={{ width: 160 }} />
                  </Form.Item>
                );
              }}
            </Form.Item>
            <Form.Item noStyle shouldUpdate>
              {({ getFieldValue }) => {
                const scopeType = getFieldValue('scopeType') as ChannelTemplateScopeType;
                if (scopeType !== 'merchant') {
                  return null;
                }
                return (
                  <Form.Item
                    label="商户 ID"
                    name="merchantId"
                    rules={[{ required: true, message: '请输入商户 ID' }]}
                  >
                    <InputNumber min={1} style={{ width: 160 }} />
                  </Form.Item>
                );
              }}
            </Form.Item>
          </Space>

          <Form.Item label="元数据配置" name="metadataConfigText">
            <TextArea autoSize={{ minRows: 3, maxRows: 8 }} />
          </Form.Item>
          <Form.Item label="敏感配置引用" name="secretRefConfigText">
            <TextArea autoSize={{ minRows: 3, maxRows: 8 }} />
          </Form.Item>
          <Form.Item label="意图契约配置" name="intentContractConfigText">
            <TextArea autoSize={{ minRows: 3, maxRows: 8 }} />
          </Form.Item>
          <Form.Item label="影响范围配置" name="impactScopeConfigText">
            <TextArea autoSize={{ minRows: 3, maxRows: 8 }} />
          </Form.Item>
          <Form.Item label="备注" name="remark">
            <TextArea autoSize={{ minRows: 2, maxRows: 4 }} maxLength={255} />
          </Form.Item>
        </Form>
      </Modal>

      <Drawer
        width={720}
        visible={detailVisible}
        destroyOnClose
        title={currentRow ? `模板详情 · ${currentRow.templateName}` : '模板详情'}
        onClose={() => {
          setDetailVisible(false);
          setCurrentRow(undefined);
        }}
      >
        {detailLoading ? null : currentRow && (
          <>
            <Alert
              showIcon
              type="warning"
              style={{ marginBottom: 16 }}
              message="Secret-safe 约束"
              description="敏感配置只展示引用信息，不应在此页面回填明文。后续如需接入密钥中心，可继续在此页补充校验与跳转。"
            />
            <Descriptions bordered column={1} size="small">
              <Descriptions.Item label="模板名称">{currentRow.templateName}</Descriptions.Item>
              <Descriptions.Item label="模板编码">{currentRow.templateCode}</Descriptions.Item>
              <Descriptions.Item label="模板类型">{templateTypeValueEnum[currentRow.templateType]?.text}</Descriptions.Item>
              <Descriptions.Item label="目标编码">{formatTargetLabel(currentRow)}</Descriptions.Item>
              <Descriptions.Item label="作用域">{scopeTypeValueEnum[currentRow.scopeType]?.text}</Descriptions.Item>
              <Descriptions.Item label="状态">
                <Tag color={statusColorMap[currentRow.status]}>{statusValueEnum[currentRow.status]?.text}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="状态说明">{currentRow.statusReason || '-'}</Descriptions.Item>
              <Descriptions.Item label="影响范围摘要">{formatImpactSummary(currentRow)}</Descriptions.Item>
              <Descriptions.Item label="影响范围样本">
                {currentRow.bindingSampleLabels && currentRow.bindingSampleLabels.length > 0
                  ? currentRow.bindingSampleLabels.join('、')
                  : '-'}
              </Descriptions.Item>
              <Descriptions.Item label="最近状态变更时间">{currentRow.lastStatusChangeTime || '-'}</Descriptions.Item>
              <Descriptions.Item label="创建信息">
                {currentRow.createBy || '-'} · {currentRow.createTime || '-'}
              </Descriptions.Item>
              <Descriptions.Item label="最近更新">
                {currentRow.updateBy || '-'} · {currentRow.updateTime || '-'}
              </Descriptions.Item>
              <Descriptions.Item label="元数据配置">
                <pre style={{ margin: 0, whiteSpace: 'pre-wrap' }}>{formatJSONText(currentRow.metadataConfig)}</pre>
              </Descriptions.Item>
              <Descriptions.Item label="敏感配置引用">
                <pre style={{ margin: 0, whiteSpace: 'pre-wrap' }}>{formatJSONText(currentRow.secretRefConfig)}</pre>
              </Descriptions.Item>
              <Descriptions.Item label="意图契约配置">
                <pre style={{ margin: 0, whiteSpace: 'pre-wrap' }}>{formatJSONText(currentRow.intentContractConfig)}</pre>
              </Descriptions.Item>
              <Descriptions.Item label="影响范围配置">
                <pre style={{ margin: 0, whiteSpace: 'pre-wrap' }}>{formatJSONText(currentRow.impactScopeConfig)}</pre>
              </Descriptions.Item>
              <Descriptions.Item label="备注">{currentRow.remark || '-'}</Descriptions.Item>
            </Descriptions>
          </>
        )}
      </Drawer>
    </PageContainer>
  );
};

export default TemplatePage;
