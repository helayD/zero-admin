import {
  CheckCircleOutlined,
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
  message,
  Modal,
  Space,
  Spin,
  Tag,
} from 'antd';
import React, { useRef, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import type { ActionType, ProColumns } from '@ant-design/pro-table';
import ProTable from '@ant-design/pro-table';
import CreateTenantDrawer from './components/CreateTenantDrawer';
import type {
  CreateTenantResult,
  TenantListItem,
  TenantStatusAction,
  TenantStatusPayload,
} from './data.d';
import {
  archiveTenant,
  disableTenant,
  enableTenant,
  queryTenantDetail,
  queryTenantList,
} from './service';

const { confirm } = Modal;

const channelOptions = [
  { label: 'App', value: 'app' },
  { label: '小程序', value: 'mini-program' },
  { label: 'H5', value: 'h5' },
  { label: '门店 POS', value: 'pos' },
];

const featureOptions = [
  { label: 'OMS 订单履约', value: 'oms' },
  { label: 'PMS 商品中心', value: 'pms' },
  { label: 'CMS 内容运营', value: 'cms' },
  { label: 'SMS 营销活动', value: 'sms' },
  { label: 'CRM 会员运营', value: 'crm' },
];

const statusMap: Record<number, { text: string; color: string }> = {
  0: { text: '已停用', color: 'default' },
  1: { text: '已启用', color: 'success' },
  2: { text: '待激活', color: 'processing' },
  3: { text: '已归档', color: 'warning' },
};

const activationMap: Record<string, { text: string; color: string }> = {
  active: { text: '已激活', color: 'success' },
  disabled: { text: '已停用', color: 'default' },
  archived: { text: '已归档', color: 'warning' },
  'pending-activation': { text: '待激活', color: 'processing' },
};

const actionCopy: Record<
  TenantStatusAction,
  { title: string; icon: React.ReactNode; service: (params: TenantStatusPayload) => Promise<any> }
> = {
  enable: {
    title: '启用',
    icon: <CheckCircleOutlined />,
    service: enableTenant,
  },
  disable: {
    title: '停用',
    icon: <PauseCircleOutlined />,
    service: disableTenant,
  },
  archive: {
    title: '归档',
    icon: <InboxOutlined />,
    service: archiveTenant,
  },
};

const readErrorMessage = (error: any, fallback: string) => {
  return error?.data?.message || error?.message || fallback;
};

const renderStatusTag = (status?: number) => {
  const meta = statusMap[status || 0] || { text: '未知', color: 'default' };
  return <Tag color={meta.color}>{meta.text}</Tag>;
};

const renderActivationTag = (status?: string) => {
  const meta = activationMap[status || 'pending-activation'] || activationMap['pending-activation'];
  return <Tag color={meta.color}>{meta.text}</Tag>;
};

const TenantPage: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const [createVisible, setCreateVisible] = useState(false);
  const [detailVisible, setDetailVisible] = useState(false);
  const [detailLoading, setDetailLoading] = useState(false);
  const [currentRow, setCurrentRow] = useState<TenantListItem>();
  const [selectedRows, setSelectedRows] = useState<TenantListItem[]>([]);
  const [lastCreated, setLastCreated] = useState<CreateTenantResult>();

  const openDetail = async (record: TenantListItem) => {
    setDetailVisible(true);
    setDetailLoading(true);
    try {
      const result = await queryTenantDetail(record.id);
      setCurrentRow(result.data);
    } catch (error) {
      message.error(readErrorMessage(error, '加载租户详情失败，请稍后重试'));
      setDetailVisible(false);
    } finally {
      setDetailLoading(false);
    }
  };

  const reloadTable = () => {
    actionRef.current?.reload?.();
    setSelectedRows([]);
  };

  const executeStatusAction = async (action: TenantStatusAction, ids: number[]) => {
    const hide = message.loading(`正在${actionCopy[action].title}租户`);
    try {
      await actionCopy[action].service({ ids });
      hide();
      message.success(`${actionCopy[action].title}成功`);
      reloadTable();
      if (currentRow && ids.includes(currentRow.id)) {
        openDetail({ ...currentRow, id: currentRow.id });
      }
      return true;
    } catch (error) {
      hide();
      message.error(
        readErrorMessage(error, `${actionCopy[action].title}失败，请先核对当前租户状态`),
      );
      return false;
    }
  };

  const showStatusConfirm = (action: TenantStatusAction, rows: TenantListItem[]) => {
    const ids = rows.map((item) => item.id);
    confirm({
      title: `确认${actionCopy[action].title}${
        ids.length > 1 ? `这 ${ids.length} 个` : '该'
      }租户吗？`,
      icon: actionCopy[action].icon,
      content:
        action === 'archive'
          ? '归档后会保留审计轨迹，但不再参与后续治理流转。'
          : action === 'disable'
          ? '停用后后台查询会在短时间内反映最新状态，管理员仍保留恢复入口。'
          : '启用后租户会进入可继续治理的状态，建议随后补充角色模板与业务域配置。',
      onOk: async () => executeStatusAction(action, ids),
    });
  };

  const columns: ProColumns<TenantListItem>[] = [
    {
      title: '租户名称',
      dataIndex: 'tenantName',
      render: (_, record) => <a onClick={() => openDetail(record)}>{record.tenantName}</a>,
    },
    {
      title: '租户标识',
      dataIndex: 'tenantCode',
    },
    {
      title: '可用渠道',
      dataIndex: 'availableChannels',
      hideInSearch: true,
      render: (_, record) => (
        <Space wrap>
          {(record.availableChannels || []).map((channel) => (
            <Tag key={channel}>{channel}</Tag>
          ))}
        </Space>
      ),
    },
    {
      title: '渠道筛选',
      dataIndex: 'channel',
      hideInTable: true,
      valueType: 'select',
      valueEnum: {
        app: { text: 'App' },
        'mini-program': { text: '小程序' },
        h5: { text: 'H5' },
        pos: { text: '门店 POS' },
      },
    },
    {
      title: '生命周期',
      dataIndex: 'status',
      valueType: 'select',
      valueEnum: {
        '-1': { text: '全部' },
        0: { text: '已停用' },
        1: { text: '已启用' },
        2: { text: '待激活' },
        3: { text: '已归档' },
      },
      render: (_, record) => renderStatusTag(record.status),
    },
    {
      title: '首个管理员',
      dataIndex: 'primaryAdminUserName',
      hideInSearch: true,
      render: (_, record) => (
        <div>
          <div>{record.primaryAdminUserName || '-'}</div>
          <div style={{ color: 'rgba(0,0,0,0.45)' }}>{record.primaryAdminMobile || '未初始化'}</div>
        </div>
      ),
    },
    {
      title: '管理员激活态',
      dataIndex: 'adminActivationStatus',
      hideInSearch: true,
      render: (_, record) => renderActivationTag(record.adminActivationStatus),
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      hideInSearch: true,
    },
    {
      title: '操作',
      valueType: 'option',
      width: 220,
      render: (_, record) => {
        const actions: React.ReactNode[] = [
          <a key="detail" onClick={() => openDetail(record)}>
            <EyeOutlined /> 详情
          </a>,
        ];
        if (record.status !== 1 && record.status !== 3) {
          actions.push(
            <a key="enable" onClick={() => showStatusConfirm('enable', [record])}>
              <CheckCircleOutlined /> 启用
            </a>,
          );
        }
        if (record.status !== 0 && record.status !== 3) {
          actions.push(
            <a key="disable" onClick={() => showStatusConfirm('disable', [record])}>
              <PauseCircleOutlined /> 停用
            </a>,
          );
        }
        if (record.status !== 3) {
          actions.push(
            <a key="archive" onClick={() => showStatusConfirm('archive', [record])}>
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
        message="Scope Context Bar"
        description="当前处于平台控制台 / 租户治理。这里的操作会直接影响租户生命周期、首个管理员 bootstrap 状态以及后续治理链路，请在同一页面内完成创建、查看和状态流转。"
        style={{ marginBottom: 16 }}
      />

      {lastCreated && (
        <Alert
          showIcon
          closable
          type="success"
          message={`租户 ${lastCreated.tenantCode} 已创建`}
          description={`下一步建议：1）确认是否立即启用租户；2）补充角色模板与菜单授权；3）继续在后续 story 中接入业务域配置。首个管理员账号 ID 为 ${lastCreated.adminUserId}。`}
          style={{ marginBottom: 16 }}
          onClose={() => setLastCreated(undefined)}
        />
      )}

      <ProTable<TenantListItem>
        headerTitle="租户治理"
        actionRef={actionRef}
        rowKey="id"
        search={{ labelWidth: 120 }}
        request={queryTenantList}
        columns={columns}
        pagination={{ pageSize: 20 }}
        rowSelection={{
          onChange: (_, rows) => setSelectedRows(rows),
        }}
        toolBarRender={() => [
          <Button key="create" type="primary" onClick={() => setCreateVisible(true)}>
            <PlusOutlined /> 创建租户
          </Button>,
        ]}
        tableAlertOptionRender={() => {
          if (selectedRows.length === 0) {
            return false;
          }
          return (
            <Space size={16}>
              <a onClick={() => showStatusConfirm('enable', selectedRows)}>批量启用</a>
              <a onClick={() => showStatusConfirm('disable', selectedRows)}>批量停用</a>
              <a onClick={() => showStatusConfirm('archive', selectedRows)}>批量归档</a>
            </Space>
          );
        }}
      />

      <CreateTenantDrawer
        visible={createVisible}
        onVisibleChange={setCreateVisible}
        channelOptions={channelOptions}
        featureOptions={featureOptions}
        onSuccess={(result) => {
          setLastCreated(result);
          setCreateVisible(false);
          reloadTable();
        }}
      />

      <Drawer
        title={currentRow ? `租户详情 · ${currentRow.tenantName}` : '租户详情'}
        width={620}
        visible={detailVisible}
        destroyOnClose
        onClose={() => {
          setDetailVisible(false);
          setCurrentRow(undefined);
        }}
      >
        <Spin spinning={detailLoading}>
          {currentRow && (
            <>
              <Alert
                showIcon
                type="warning"
                message="Recovery-first Feedback"
                description="如果当前租户仍处于待激活或停用状态，建议先完成生命周期恢复，再继续做角色、菜单和业务域开通，避免把问题留到后续流程里。"
                style={{ marginBottom: 16 }}
              />

              <Descriptions bordered column={1} size="small">
                <Descriptions.Item label="租户名称">{currentRow.tenantName}</Descriptions.Item>
                <Descriptions.Item label="租户标识">{currentRow.tenantCode}</Descriptions.Item>
                <Descriptions.Item label="生命周期">
                  {renderStatusTag(currentRow.status)}
                </Descriptions.Item>
                <Descriptions.Item label="状态说明">
                  {currentRow.statusReason || '-'}
                </Descriptions.Item>
                <Descriptions.Item label="可用渠道">
                  <Space wrap>
                    {(currentRow.availableChannels || []).map((channel) => (
                      <Tag key={channel}>{channel}</Tag>
                    ))}
                  </Space>
                </Descriptions.Item>
                <Descriptions.Item label="业务开关">
                  <Space wrap>
                    {(currentRow.featureFlags || []).length > 0
                      ? currentRow.featureFlags.map((flag) => <Tag key={flag}>{flag}</Tag>)
                      : '-'}
                  </Space>
                </Descriptions.Item>
                <Descriptions.Item label="数据保留策略">
                  {currentRow.dataRetentionDays} 天
                </Descriptions.Item>
                <Descriptions.Item label="联系人">
                  {currentRow.contactName} / {currentRow.contactMobile}
                </Descriptions.Item>
                <Descriptions.Item label="联系邮箱">
                  {currentRow.contactEmail || '-'}
                </Descriptions.Item>
                <Descriptions.Item label="首个管理员">
                  {currentRow.primaryAdminUserName || '-'} / {currentRow.primaryAdminMobile || '-'}
                </Descriptions.Item>
                <Descriptions.Item label="管理员激活态">
                  {renderActivationTag(currentRow.adminActivationStatus)}
                </Descriptions.Item>
                <Descriptions.Item label="创建信息">
                  {currentRow.createdBy || '-'} · {currentRow.createdAt || '-'}
                </Descriptions.Item>
                <Descriptions.Item label="最近更新">
                  {currentRow.updatedBy || '-'} · {currentRow.updatedAt || '-'}
                </Descriptions.Item>
              </Descriptions>

              <div style={{ marginTop: 16 }}>
                <Space wrap>
                  {currentRow.status !== 1 && currentRow.status !== 3 && (
                    <Button
                      type="primary"
                      onClick={() => showStatusConfirm('enable', [currentRow])}
                    >
                      立即启用
                    </Button>
                  )}
                  {currentRow.status !== 0 && currentRow.status !== 3 && (
                    <Button onClick={() => showStatusConfirm('disable', [currentRow])}>停用</Button>
                  )}
                  {currentRow.status !== 3 && (
                    <Button danger onClick={() => showStatusConfirm('archive', [currentRow])}>
                      归档
                    </Button>
                  )}
                </Space>
              </div>
            </>
          )}
        </Spin>
      </Drawer>
    </PageContainer>
  );
};

export default TenantPage;
