import {
  CheckCircleOutlined,
  CloseCircleOutlined,
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
  message,
  Space,
  Spin,
  Tag,
} from 'antd';
import React, { useRef, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import type { ActionType, ProColumns } from '@ant-design/pro-table';
import ProTable from '@ant-design/pro-table';
import ActionReasonModal from './components/ActionReasonModal';
import CreateMerchantDrawer from './components/CreateMerchantDrawer';
import type {
  CreateMerchantResult,
  MerchantListItem,
  MerchantReviewAction,
  MerchantReviewPayload,
  MerchantStatusAction,
  MerchantStatusPayload,
} from './data.d';
import {
  approveMerchant,
  archiveMerchant,
  disableMerchant,
  enableMerchant,
  queryMerchantDetail,
  queryMerchantList,
  rejectMerchant,
  requestMerchantMaterial,
} from './service';

type MerchantActionKind = MerchantReviewAction | MerchantStatusAction;

const channelOptions = [
  { label: 'App', value: 'app' },
  { label: '小程序', value: 'mini-program' },
  { label: 'H5', value: 'h5' },
  { label: '门店 POS', value: 'pos' },
];

const capabilityOptions = [
  { label: 'OMS 订单履约', value: 'oms' },
  { label: 'PMS 商品中心', value: 'pms' },
  { label: 'CMS 内容运营', value: 'cms' },
  { label: 'SMS 营销活动', value: 'sms' },
  { label: 'CRM 会员运营', value: 'crm' },
];

const reviewStatusMap: Record<number, { text: string; color: string }> = {
  0: { text: '待审核', color: 'processing' },
  1: { text: '补充材料', color: 'warning' },
  2: { text: '已驳回', color: 'error' },
  3: { text: '已通过', color: 'success' },
};

const businessStatusMap: Record<number, { text: string; color: string }> = {
  0: { text: '待激活', color: 'processing' },
  1: { text: '已启用', color: 'success' },
  2: { text: '已停用', color: 'default' },
  3: { text: '已归档', color: 'warning' },
};

const tenantStatusMap: Record<number, { text: string; color: string }> = {
  0: { text: '已停用', color: 'default' },
  1: { text: '已启用', color: 'success' },
  2: { text: '待激活', color: 'processing' },
  3: { text: '已归档', color: 'warning' },
};

const actionMeta: Record<
  MerchantActionKind,
  {
    title: string;
    confirmText: string;
    description: string;
    placeholder: string;
    danger?: boolean;
    requireReason?: boolean;
  }
> = {
  approve: {
    title: '审核通过',
    confirmText: '确认通过',
    description: '通过后商户主体会进入待激活状态，后续可继续执行启用、菜单模板分配和商户管理员绑定。',
    placeholder: '可填写审核结论，如租户归属确认、能力包核验结果等（可选）',
  },
  reject: {
    title: '驳回申请',
    confirmText: '确认驳回',
    description: '驳回会保留当前申请与审计轨迹。请明确记录拒绝原因，避免后续治理链路缺少上下文。',
    placeholder: '请填写驳回原因，例如资质缺失、租户归属不合法等',
    danger: true,
    requireReason: true,
  },
  material: {
    title: '要求补充材料',
    confirmText: '确认退回',
    description: '该动作会把申请退回到“补充材料”状态，建议给出清晰的补充项清单，便于下一次审核直接收口。',
    placeholder: '请填写需要补充的材料或需要修正的字段',
    requireReason: true,
  },
  enable: {
    title: '启用商户',
    confirmText: '确认启用',
    description: '启用会同步恢复商户级 scope 激活态。为确保权限和缓存刷新一致，建议商户管理员重新登录。',
    placeholder: '可填写启用说明，例如开业时间、能力包生效批次（可选）',
  },
  disable: {
    title: '停用商户',
    confirmText: '确认停用',
    description: '停用会同步回收商户后台访问与关键经营能力，并清理权限缓存。请说明原因，便于后续恢复。',
    placeholder: '请填写停用原因，例如资质异常、风控冻结、租户暂停经营等',
    danger: true,
    requireReason: true,
  },
  archive: {
    title: '归档商户',
    confirmText: '确认归档',
    description: '归档后主体会退出后续治理流转，但审计记录会保留。请仅在明确结束该主体生命周期时使用。',
    placeholder: '请填写归档原因，例如申请作废、主体清退完成等',
    danger: true,
    requireReason: true,
  },
};

const readErrorMessage = (error: any, fallback: string) => {
  return error?.data?.message || error?.message || fallback;
};

const renderStatusTag = (metaMap: Record<number, { text: string; color: string }>, status?: number) => {
  const meta = metaMap[status || 0] || { text: '未知', color: 'default' };
  return <Tag color={meta.color}>{meta.text}</Tag>;
};

const MerchantPage: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const [createVisible, setCreateVisible] = useState(false);
  const [detailVisible, setDetailVisible] = useState(false);
  const [detailLoading, setDetailLoading] = useState(false);
  const [currentRow, setCurrentRow] = useState<MerchantListItem>();
  const [selectedRows, setSelectedRows] = useState<MerchantListItem[]>([]);
  const [lastCreated, setLastCreated] = useState<CreateMerchantResult>();
  const [actionKind, setActionKind] = useState<MerchantActionKind>();
  const [actionRows, setActionRows] = useState<MerchantListItem[]>([]);

  const openDetail = async (record: MerchantListItem) => {
    setDetailVisible(true);
    setDetailLoading(true);
    try {
      const result = await queryMerchantDetail(record.id);
      setCurrentRow(result.data);
    } catch (error) {
      message.error(readErrorMessage(error, '加载商户详情失败，请稍后重试'));
      setDetailVisible(false);
    } finally {
      setDetailLoading(false);
    }
  };

  const reloadTable = () => {
    actionRef.current?.reload?.();
    setSelectedRows([]);
  };

  const closeActionModal = () => {
    setActionKind(undefined);
    setActionRows([]);
  };

  const openActionModal = (kind: MerchantActionKind, rows: MerchantListItem[]) => {
    setActionKind(kind);
    setActionRows(rows);
  };

  const submitAction = async (reason: string) => {
    if (!actionKind || actionRows.length === 0) {
      return false;
    }

    const ids = actionRows.map((item) => item.id);
    const hide = message.loading(`正在执行${actionMeta[actionKind].title}`);
    try {
      if (actionKind === 'approve' || actionKind === 'reject' || actionKind === 'material') {
        const payload: MerchantReviewPayload = { ids, reviewReason: reason || undefined };
        if (actionKind === 'approve') {
          await approveMerchant(payload);
        } else if (actionKind === 'reject') {
          await rejectMerchant(payload);
        } else {
          await requestMerchantMaterial(payload);
        }
      } else {
        const payload: MerchantStatusPayload = { ids, statusReason: reason || undefined };
        if (actionKind === 'enable') {
          await enableMerchant(payload);
        } else if (actionKind === 'disable') {
          await disableMerchant(payload);
        } else {
          await archiveMerchant(payload);
        }
      }

      hide();
      message.success(`${actionMeta[actionKind].title}成功`);
      closeActionModal();
      reloadTable();
      if (currentRow && ids.includes(currentRow.id)) {
        await openDetail({ ...currentRow, id: currentRow.id });
      }
      return true;
    } catch (error) {
      hide();
      message.error(readErrorMessage(error, `${actionMeta[actionKind].title}失败，请核对当前状态后重试`));
      return false;
    }
  };

  const columns: ProColumns<MerchantListItem>[] = [
    {
      title: '商户名称',
      dataIndex: 'merchantName',
      render: (_, record) => <a onClick={() => openDetail(record)}>{record.merchantName}</a>,
    },
    {
      title: '商户编码',
      dataIndex: 'merchantCode',
    },
    {
      title: '归属租户',
      dataIndex: 'tenantName',
      render: (_, record) => (
        <div>
          <div>{record.tenantName || '-'}</div>
          <div style={{ color: 'rgba(0,0,0,0.45)' }}>{record.tenantCode || '-'}</div>
        </div>
      ),
    },
    {
      title: '租户 ID',
      dataIndex: 'tenantId',
      hideInTable: true,
      fieldProps: { precision: 0 },
      valueType: 'digit',
    },
    {
      title: '审核状态',
      dataIndex: 'reviewStatus',
      valueType: 'select',
      valueEnum: {
        '-1': { text: '全部' },
        0: { text: '待审核' },
        1: { text: '补充材料' },
        2: { text: '已驳回' },
        3: { text: '已通过' },
      },
      render: (_, record) => renderStatusTag(reviewStatusMap, record.reviewStatus),
    },
    {
      title: '经营状态',
      dataIndex: 'businessStatus',
      valueType: 'select',
      valueEnum: {
        '-1': { text: '全部' },
        0: { text: '待激活' },
        1: { text: '已启用' },
        2: { text: '已停用' },
        3: { text: '已归档' },
      },
      render: (_, record) => renderStatusTag(businessStatusMap, record.businessStatus),
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
      title: '能力包筛选',
      dataIndex: 'capabilityFlag',
      hideInTable: true,
      valueType: 'select',
      valueEnum: {
        oms: { text: 'OMS 订单履约' },
        pms: { text: 'PMS 商品中心' },
        cms: { text: 'CMS 内容运营' },
        sms: { text: 'SMS 营销活动' },
        crm: { text: 'CRM 会员运营' },
      },
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
      title: '能力包',
      dataIndex: 'capabilityFlags',
      hideInSearch: true,
      render: (_, record) => (
        <Space wrap>
          {(record.capabilityFlags || []).map((capability) => (
            <Tag key={capability} color="blue">
              {capability}
            </Tag>
          ))}
        </Space>
      ),
    },
    {
      title: '处理人 / 时间',
      dataIndex: 'reviewedByName',
      hideInSearch: true,
      render: (_, record) => (
        <div>
          <div>{record.reviewedByName || '-'}</div>
          <div style={{ color: 'rgba(0,0,0,0.45)' }}>{record.reviewedAt || '-'}</div>
        </div>
      ),
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
        ];
        if (record.reviewStatus !== 3 && record.businessStatus !== 3) {
          actions.push(
            <a key="approve" onClick={() => openActionModal('approve', [record])}>
              <CheckCircleOutlined /> 通过
            </a>,
          );
          actions.push(
            <a key="material" onClick={() => openActionModal('material', [record])}>
              <EditOutlined /> 补材
            </a>,
          );
          actions.push(
            <a key="reject" onClick={() => openActionModal('reject', [record])}>
              <CloseCircleOutlined /> 驳回
            </a>,
          );
        }
        if (record.reviewStatus === 3 && record.businessStatus !== 1 && record.businessStatus !== 3) {
          actions.push(
            <a key="enable" onClick={() => openActionModal('enable', [record])}>
              <CheckCircleOutlined /> 启用
            </a>,
          );
        }
        if (record.businessStatus === 1) {
          actions.push(
            <a key="disable" onClick={() => openActionModal('disable', [record])}>
              <PauseCircleOutlined /> 停用
            </a>,
          );
        }
        if (record.businessStatus !== 3) {
          actions.push(
            <a key="archive" onClick={() => openActionModal('archive', [record])}>
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
        description="当前处于平台控制台 / 商户治理。这里的审核、启停和归档会直接影响商户主体状态、后台访问权限、能力包生效与后续菜单模板/作用域分配。"
        style={{ marginBottom: 16 }}
      />

      <Alert
        showIcon
        type="info"
        message="Consequence Preview"
        description="审核通过只代表商户主体合规进入待激活阶段；真正开放后台访问需要继续执行启用动作。停用时会同步回收商户级访问缓存，恢复后建议管理员重新登录。"
        style={{ marginBottom: 16 }}
      />

      {lastCreated && (
        <Alert
          showIcon
          closable
          type="success"
          message={`商户 ${lastCreated.merchantCode} 已创建`}
          description="下一步建议：1）核对租户归属和能力包；2）执行审核通过或要求补充材料；3）若已有商户管理员，启用前确认商户级 scope 绑定是否齐备。"
          style={{ marginBottom: 16 }}
          onClose={() => setLastCreated(undefined)}
        />
      )}

      {selectedRows.length > 0 && (
        <Alert
          showIcon
          type="warning"
          message={`Batch Action Dock · 已选择 ${selectedRows.length} 个商户`}
          description="批量审核/启停会同时影响商户后台访问、能力包生效和后续治理入口。请确认这些商户属于同一批次决策，再继续提交。"
          style={{ marginBottom: 16 }}
        />
      )}

      <ProTable<MerchantListItem>
        headerTitle="商户治理"
        actionRef={actionRef}
        rowKey="id"
        search={{ labelWidth: 120 }}
        request={queryMerchantList}
        columns={columns}
        pagination={{ pageSize: 20 }}
        rowSelection={{
          onChange: (_, rows) => setSelectedRows(rows),
        }}
        toolBarRender={() => [
          <Button key="create" type="primary" onClick={() => setCreateVisible(true)}>
            <PlusOutlined /> 创建申请
          </Button>,
        ]}
        tableAlertOptionRender={() => {
          if (selectedRows.length === 0) {
            return false;
          }
          return (
            <Space size={16}>
              <a onClick={() => openActionModal('approve', selectedRows)}>批量通过</a>
              <a onClick={() => openActionModal('material', selectedRows)}>批量补材</a>
              <a onClick={() => openActionModal('reject', selectedRows)}>批量驳回</a>
              <a onClick={() => openActionModal('enable', selectedRows)}>批量启用</a>
              <a onClick={() => openActionModal('disable', selectedRows)}>批量停用</a>
              <a onClick={() => openActionModal('archive', selectedRows)}>批量归档</a>
            </Space>
          );
        }}
      />

      <CreateMerchantDrawer
        visible={createVisible}
        onVisibleChange={setCreateVisible}
        channelOptions={channelOptions}
        capabilityOptions={capabilityOptions}
        onSuccess={(result) => {
          setLastCreated(result);
          setCreateVisible(false);
          reloadTable();
        }}
      />

      {actionKind && (
        <ActionReasonModal
          visible={!!actionKind}
          title={`${actionMeta[actionKind].title}${actionRows.length > 1 ? `（${actionRows.length} 个商户）` : ''}`}
          description={actionMeta[actionKind].description}
          confirmText={actionMeta[actionKind].confirmText}
          placeholder={actionMeta[actionKind].placeholder}
          danger={actionMeta[actionKind].danger}
          requireReason={actionMeta[actionKind].requireReason}
          onCancel={closeActionModal}
          onSubmit={submitAction}
        />
      )}

      <Drawer
        title={currentRow ? `商户详情 · ${currentRow.merchantName}` : '商户详情'}
        width={720}
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
                description="若当前商户已停用，后台访问与能力包缓存已经被回收。恢复经营前建议先确认租户状态、商户管理员绑定和菜单模板分配是否仍然完整。"
                style={{ marginBottom: 16 }}
              />

              <Descriptions bordered column={1} size="small">
                <Descriptions.Item label="商户名称">{currentRow.merchantName}</Descriptions.Item>
                <Descriptions.Item label="商户编码">{currentRow.merchantCode}</Descriptions.Item>
                <Descriptions.Item label="归属租户">
                  {currentRow.tenantName || '-'} / {currentRow.tenantCode || '-'}
                </Descriptions.Item>
                <Descriptions.Item label="租户状态">
                  {renderStatusTag(tenantStatusMap, currentRow.tenantStatus)}
                </Descriptions.Item>
                <Descriptions.Item label="审核状态">
                  {renderStatusTag(reviewStatusMap, currentRow.reviewStatus)}
                </Descriptions.Item>
                <Descriptions.Item label="审核说明">
                  {currentRow.reviewReason || '-'}
                </Descriptions.Item>
                <Descriptions.Item label="审核处理人">
                  {currentRow.reviewedByName || '-'} / {currentRow.reviewedAt || '-'}
                </Descriptions.Item>
                <Descriptions.Item label="经营状态">
                  {renderStatusTag(businessStatusMap, currentRow.businessStatus)}
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
                <Descriptions.Item label="能力包">
                  <Space wrap>
                    {(currentRow.capabilityFlags || []).map((flag) => (
                      <Tag key={flag} color="blue">
                        {flag}
                      </Tag>
                    ))}
                  </Space>
                </Descriptions.Item>
                <Descriptions.Item label="联系人">
                  {currentRow.contactName} / {currentRow.contactMobile}
                </Descriptions.Item>
                <Descriptions.Item label="联系邮箱">
                  {currentRow.contactEmail || '-'}
                </Descriptions.Item>
                <Descriptions.Item label="商户管理员用户 ID">
                  {currentRow.primaryAdminUserId || '-'}
                </Descriptions.Item>
                <Descriptions.Item label="可见范围提示">
                  {currentRow.visibleScopeHint || '-'}
                </Descriptions.Item>
                <Descriptions.Item label="下一步动作">
                  <Space wrap>
                    {(currentRow.nextActions || []).map((action) => (
                      <Tag key={action} color="gold">
                        {action}
                      </Tag>
                    ))}
                  </Space>
                </Descriptions.Item>
                <Descriptions.Item label="备注">{currentRow.remark || '-'}</Descriptions.Item>
                <Descriptions.Item label="创建信息">
                  {currentRow.createdBy || '-'} · {currentRow.createdAt || '-'}
                </Descriptions.Item>
                <Descriptions.Item label="最近更新">
                  {currentRow.updatedBy || '-'} · {currentRow.updatedAt || '-'}
                </Descriptions.Item>
              </Descriptions>

              <div style={{ marginTop: 16 }}>
                <Space wrap>
                  {currentRow.reviewStatus !== 3 && currentRow.businessStatus !== 3 && (
                    <>
                      <Button type="primary" onClick={() => openActionModal('approve', [currentRow])}>
                        审核通过
                      </Button>
                      <Button onClick={() => openActionModal('material', [currentRow])}>
                        要求补材
                      </Button>
                      <Button danger onClick={() => openActionModal('reject', [currentRow])}>
                        驳回
                      </Button>
                    </>
                  )}
                  {currentRow.reviewStatus === 3 &&
                    currentRow.businessStatus !== 1 &&
                    currentRow.businessStatus !== 3 && (
                      <Button type="primary" onClick={() => openActionModal('enable', [currentRow])}>
                        立即启用
                      </Button>
                    )}
                  {currentRow.businessStatus === 1 && (
                    <Button onClick={() => openActionModal('disable', [currentRow])}>停用</Button>
                  )}
                  {currentRow.businessStatus !== 3 && (
                    <Button danger onClick={() => openActionModal('archive', [currentRow])}>
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

export default MerchantPage;
