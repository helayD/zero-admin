import {
  PlusOutlined,
  ExclamationCircleOutlined,
  DeleteOutlined,
  EditOutlined,
} from '@ant-design/icons';
import { Alert, Button, Divider, Image, message, Modal, Switch, Tag } from 'antd';
import React, { useState, useRef } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import ProTable from '@ant-design/pro-table';
import type { ProColumns, ActionType } from '@ant-design/pro-table';
import type { CardTemplateListItem } from './data.d';
import { ContentAuditStatusOptions } from './data.d';
import {
  queryCardTemplateList,
  deleteCardTemplate,
  updateCardTemplateStatus,
  checkCardTemplateUsage,
} from './service';
import TemplateForm from './components/TemplateForm';
import GovernanceScopeBar from '@/pages/system/components/GovernanceScopeBar';
import {
  buildGovernanceScopeLabel,
  defaultGovernanceScope,
  type GovernanceScopeValue,
  toGovernancePayload,
} from '@/pages/system/components/governance';

const { confirm } = Modal;

const renderStatusTag = (status: number) =>
  status === 1 ? <Tag color="green">启用</Tag> : <Tag color="red">禁用</Tag>;

const renderDisplayStatusTag = (s: number) =>
  s === 1 ? <Tag color="blue">上架</Tag> : <Tag>下架</Tag>;

const renderContentAuditTag = (s: number) => {
  const map: Record<number, { color: string; text: string }> = {
    0: { color: 'default', text: '待审' },
    1: { color: 'gold', text: '审核中' },
    2: { color: 'green', text: '通过' },
    3: { color: 'red', text: '驳回' },
  };
  const item = map[s] || { color: 'default', text: String(s) };
  return <Tag color={item.color}>{item.text}</Tag>;
};

const CardTemplateList: React.FC = () => {
  const [showForm, setShowForm] = useState(false);
  const actionRef = useRef<ActionType>();
  const [currentRow, setCurrentRow] = useState<CardTemplateListItem>();
  const [scope, setScope] = useState<GovernanceScopeValue>(defaultGovernanceScope);

  const handleDelete = async (record: CardTemplateListItem) => {
    // Story 10.10 修复 M4: checkUsage 失败时弹错而非静默继续。后端 H1 拦截后仍会拒绝，但前端需给出明确反馈。
    try {
      const usage = await checkCardTemplateUsage(record.id, toGovernancePayload(scope));
      if (usage.data?.canDelete === false) {
        Modal.warning({
          title: '无法删除',
          content:
            usage.data?.message ||
            `该模板被 ${usage.data?.refRuleCount || 0} 条发卡规则引用，无法删除`,
        });
        return;
      }
    } catch (error) {
      Modal.error({
        title: '引用检查失败',
        content: '无法验证模板是否被发卡规则引用，为避免误操作已阻止删除。请检查网络后重试。',
      });
      return;
    }

    confirm({
      title: '是否删除该卡片模板?',
      icon: <ExclamationCircleOutlined />,
      content: `模板：${record.templateName} (${record.templateCode})。删除后不可恢复，请确认。`,
      onOk() {
        return deleteCardTemplate(record.id, toGovernancePayload(scope)).then(() => {
          message.success('删除成功，即将刷新');
          actionRef.current?.reload?.();
        });
      },
    });
  };

  const handleStatusChange = async (record: CardTemplateListItem, checked: boolean) => {
    const newStatus = checked ? 1 : 0;
    if (newStatus === 0) {
      // Story 10.10 修复 M4: checkUsage 失败时弹错。
      try {
        const usage = await checkCardTemplateUsage(record.id, toGovernancePayload(scope));
        if (usage.data?.canDisable === false) {
          Modal.warning({
            title: '无法禁用',
            content:
              usage.data?.message ||
              `该模板被 ${usage.data?.refRuleCount || 0} 条发卡规则引用，无法禁用`,
          });
          return;
        }
      } catch (error) {
        Modal.error({
          title: '引用检查失败',
          content: '无法验证模板是否被发卡规则引用，为避免误操作已阻止禁用。请检查网络后重试。',
        });
        return;
      }
    }

    try {
      await updateCardTemplateStatus({
        id: record.id,
        status: newStatus,
        ...toGovernancePayload(scope),
      });
      message.success(`模板已${checked ? '启用' : '禁用'}`);
      actionRef.current?.reload?.();
    } catch (error) {
      message.error('状态更新失败');
    }
  };

  const columns: ProColumns<CardTemplateListItem>[] = [
    { title: '模板ID', dataIndex: 'id', hideInSearch: true, width: 80 },
    {
      title: '模板编码',
      dataIndex: 'templateCode',
    },
    {
      title: '模板名称',
      dataIndex: 'templateName',
    },
    {
      title: '卡面',
      dataIndex: 'cardFaceImage',
      hideInSearch: true,
      width: 90,
      render: (_, entity) =>
        entity.cardFaceImage ? (
          <Image src={entity.cardFaceImage} width={64} height={64} style={{ objectFit: 'cover' }} />
        ) : (
          '-'
        ),
    },
    {
      title: '稀有度',
      dataIndex: 'rarity',
    },
    {
      title: '发行上限',
      dataIndex: 'issueLimit',
      hideInSearch: true,
      render: (_, e) => (e.issueLimit > 0 ? e.issueLimit : '不限'),
    },
    {
      title: '展示状态',
      dataIndex: 'displayStatus',
      valueType: 'select',
      valueEnum: { 0: { text: '下架' }, 1: { text: '上架' } },
      render: (_, e) => renderDisplayStatusTag(e.displayStatus),
    },
    {
      title: '内容审核',
      dataIndex: 'contentAuditStatus',
      valueType: 'select',
      valueEnum: ContentAuditStatusOptions.reduce((acc: any, item) => {
        acc[item.value] = { text: item.label };
        return acc;
      }, {}),
      render: (_, e) => renderContentAuditTag(e.contentAuditStatus),
    },
    {
      title: '引用规则数',
      dataIndex: 'refRuleCount',
      hideInSearch: true,
    },
    {
      title: '启停',
      dataIndex: 'status',
      hideInSearch: true,
      render: (_, record) => (
        <>
          {renderStatusTag(record.status)}
          <Divider type="vertical" />
          <Switch
            size="small"
            checked={record.status === 1}
            onChange={(c) => handleStatusChange(record, c)}
          />
        </>
      ),
    },
    {
      title: '更新时间',
      dataIndex: 'updateTime',
      hideInSearch: true,
      valueType: 'dateTime',
    },
    {
      title: '操作',
      dataIndex: 'option',
      valueType: 'option',
      width: 160,
      render: (_, record) => (
        <>
          <a
            onClick={() => {
              setCurrentRow(record);
              setShowForm(true);
            }}
          >
            <EditOutlined /> 编辑
          </a>
          <Divider type="vertical" />
          <a style={{ color: '#ff4d4f' }} onClick={() => handleDelete(record)}>
            <DeleteOutlined /> 删除
          </a>
        </>
      ),
    },
  ];

  return (
    <PageContainer>
      <GovernanceScopeBar
        value={scope}
        onChange={(nextScope) => {
          setScope(nextScope);
          actionRef.current?.reload?.();
        }}
        entityLabel="卡片模板"
        style={{ marginBottom: 16 }}
      />
      <Alert
        showIcon
        type="info"
        style={{ marginBottom: 16 }}
        message={`当前查询范围：${buildGovernanceScopeLabel(scope)}`}
        description="管理提货卡卡片模板，包括模板编码、卡面、稀有度、发行上限、展示状态与内容审核状态。新增/启停/删除前会校验是否被发卡规则引用。"
      />
      <ProTable<CardTemplateListItem>
        headerTitle="卡片模板管理"
        actionRef={actionRef}
        rowKey="id"
        search={{ labelWidth: 120 }}
        toolBarRender={() => [
          <Button
            type="primary"
            key="primary"
            onClick={() => {
              setCurrentRow(undefined);
              setShowForm(true);
            }}
          >
            <PlusOutlined /> 新增模板
          </Button>,
        ]}
        request={(params) =>
          queryCardTemplateList({
            ...params,
            ...toGovernancePayload(scope),
          })
        }
        columns={columns}
        pagination={{ pageSize: 10 }}
      />

      <TemplateForm
        visible={showForm}
        record={currentRow}
        scope={scope}
        onCancel={() => {
          setShowForm(false);
          setCurrentRow(undefined);
        }}
        onSuccess={() => {
          setShowForm(false);
          setCurrentRow(undefined);
          actionRef.current?.reload();
        }}
      />
    </PageContainer>
  );
};

export default CardTemplateList;
