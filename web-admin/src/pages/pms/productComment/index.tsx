import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  EyeOutlined,
  RollbackOutlined,
  StopOutlined,
} from '@ant-design/icons';
import {
  Button,
  Descriptions,
  Divider,
  Drawer,
  Form,
  Image,
  Input,
  message,
  Modal,
  Radio,
  Space,
  Tag,
  Timeline,
  Typography,
} from 'antd';
import React, { useMemo, useRef, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import type { ActionType, ProColumns } from '@ant-design/pro-table';
import ProTable from '@ant-design/pro-table';
import type {
  CommentAuditLogItem,
  CommentDetailData,
  CommentReplayItem,
  CommentListItem,
} from './data.d';
import {
  auditComment,
  batchUpdateComment,
  handleCommentAppeal,
  queryCommentDetail,
  queryCommentList,
  restoreComment,
} from './service';
import GovernanceScopeBar from '@/pages/system/components/GovernanceScopeBar';
import {
  toGovernancePayload,
  type GovernanceScopeValue,
} from '@/pages/system/components/governance';

const { Paragraph, Text } = Typography;

type DetailState = CommentDetailData & {
  replays?: CommentReplayItem[];
  auditLogs?: CommentAuditLogItem[];
};

type AuditModalState = {
  open: boolean;
  record?: CommentListItem | DetailState;
  auditStatus: number;
};

type AppealModalState = {
  open: boolean;
  record?: CommentListItem | DetailState;
  appealStatus: number;
};

const auditStatusValueEnum = {
  '-1': { text: '全部' },
  '0': { text: '待审核' },
  '1': { text: '审核通过' },
  '2': { text: '审核拒绝' },
  '3': { text: '已屏蔽' },
};

const hiddenValueEnum = {
  '-1': { text: '全部' },
  '0': { text: '显示' },
  '1': { text: '屏蔽' },
};

const renderAuditStatusTag = (status: number) => {
  switch (status) {
    case 0:
      return <Tag color="warning">待审核</Tag>;
    case 1:
      return <Tag color="success">审核通过</Tag>;
    case 2:
      return <Tag color="error">审核拒绝</Tag>;
    case 3:
      return <Tag>已屏蔽</Tag>;
    default:
      return <Tag>未知</Tag>;
  }
};

const renderHiddenTag = (hidden: number) =>
  hidden === 1 ? <Tag color="error">已屏蔽</Tag> : <Tag color="processing">显示中</Tag>;

const renderAppealStatusTag = (status: number) => {
  switch (status) {
    case 1:
      return <Tag color="warning">申诉中</Tag>;
    case 2:
      return <Tag color="success">申诉通过</Tag>;
    case 3:
      return <Tag color="error">申诉驳回</Tag>;
    default:
      return <Tag>未申诉</Tag>;
  }
};

const renderStarRating = (star: number) => (
  <Space size={2}>
    {Array.from({ length: 5 }, (_, index) => (
      <span key={index} style={{ color: index < star ? '#FF9900' : '#D9D9D9', fontSize: 14 }}>
        ★
      </span>
    ))}
  </Space>
);

const ProductCommentPage: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const [scope, setScope] = useState<GovernanceScopeValue>({});
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [detailLoading, setDetailLoading] = useState(false);
  const [selectedComment, setSelectedComment] = useState<DetailState>();
  const [auditModal, setAuditModal] = useState<AuditModalState>({ open: false, auditStatus: 1 });
  const [appealModal, setAppealModal] = useState<AppealModalState>({ open: false, appealStatus: 2 });
  const [auditForm] = Form.useForm<{ auditRemark?: string }>();
  const [appealForm] = Form.useForm<{ appealReply: string }>();

  const scopePayload = useMemo(() => toGovernancePayload(scope), [scope]);

  const reloadTable = () => {
    actionRef.current?.reload();
  };

  const openDetail = async (id: string) => {
    setDrawerVisible(true);
    setDetailLoading(true);
    try {
      const res = await queryCommentDetail({ id }, scopePayload);
      if (res.code === '000000') {
        setSelectedComment({
          ...res.data,
          replays: res.replays || [],
          auditLogs: res.auditLogs || [],
        });
      } else {
        message.error(res.message || '获取评价详情失败');
      }
    } catch (error) {
      message.error('获取评价详情失败');
    } finally {
      setDetailLoading(false);
    }
  };

  const refreshCurrentDetail = () => {
    if (selectedComment?.id) {
      openDetail(selectedComment.id);
    }
  };

  const handleBatchAction = async (showStatus: number, actionText: string) => {
    if (selectedRowKeys.length === 0) {
      message.warning('请先选择要操作的评价');
      return;
    }
    Modal.confirm({
      title: `确认${actionText}`,
      content: `确定要${actionText}选中的 ${selectedRowKeys.length} 条评价吗？`,
      okText: '确认',
      cancelText: '取消',
      onOk: async () => {
        const res = await batchUpdateComment(selectedRowKeys as string[], showStatus, scopePayload);
        if (res.code === '000000') {
          message.success(res.message || `${actionText}成功`);
          setSelectedRowKeys([]);
          reloadTable();
          refreshCurrentDetail();
        } else {
          message.error(res.message || `${actionText}失败`);
        }
      },
    });
  };

  const openAuditModal = (record: CommentListItem | DetailState, auditStatus: number) => {
    setAuditModal({ open: true, record, auditStatus });
    auditForm.setFieldsValue({ auditRemark: record.auditStatus === 2 ? record.auditRemark : '' });
  };

  const submitAudit = async () => {
    if (!auditModal.record) {
      return;
    }
    const values = await auditForm.validateFields();
    const res = await auditComment(
      {
        id: auditModal.record.id,
        auditStatus: auditModal.auditStatus,
        auditRemark: values.auditRemark,
      },
      scopePayload,
    );
    if (res.code === '000000') {
      message.success(res.message || '操作成功');
      setAuditModal({ open: false, auditStatus: 1 });
      auditForm.resetFields();
      reloadTable();
      refreshCurrentDetail();
      return;
    }
    message.error(res.message || '操作失败');
  };

  const confirmRestore = (record: CommentListItem | DetailState) => {
    Modal.confirm({
      title: '确认恢复评价',
      content: '恢复后，该评价会重新进入前台可见范围。',
      okText: '确认恢复',
      cancelText: '取消',
      onOk: async () => {
        const res = await restoreComment(record.id, scopePayload);
        if (res.code === '000000') {
          message.success(res.message || '恢复成功');
          reloadTable();
          refreshCurrentDetail();
        } else {
          message.error(res.message || '恢复失败');
        }
      },
    });
  };

  const openAppealModal = (record: CommentListItem | DetailState, appealStatus = 2) => {
    setAppealModal({ open: true, record, appealStatus });
    appealForm.setFieldsValue({ appealReply: record.appealReply || '' });
  };

  const submitAppeal = async () => {
    if (!appealModal.record) {
      return;
    }
    const values = await appealForm.validateFields();
    const res = await handleCommentAppeal(
      {
        id: appealModal.record.id,
        appealStatus: appealModal.appealStatus,
        appealReply: values.appealReply,
      },
      scopePayload,
    );
    if (res.code === '000000') {
      message.success(res.message || '申诉处理成功');
      setAppealModal({ open: false, appealStatus: 2 });
      appealForm.resetFields();
      reloadTable();
      refreshCurrentDetail();
      return;
    }
    message.error(res.message || '申诉处理失败');
  };

  const columns: ProColumns<CommentListItem>[] = [
    {
      title: '评价ID',
      dataIndex: 'id',
      width: 160,
      copyable: true,
      hideInSearch: true,
    },
    {
      title: '商品名称',
      dataIndex: 'productName',
      ellipsis: true,
      width: 220,
      render: (_, record) => (
        <Space direction="vertical" size={2}>
          <Text>{record.productName || record.productAttribute || '商品评价'}</Text>
          <Text type="secondary" style={{ fontSize: 12 }}>
            商品ID: {record.productId}
          </Text>
        </Space>
      ),
    },
    {
      title: '会员昵称',
      dataIndex: 'memberNickName',
      width: 140,
      hideInSearch: true,
    },
    {
      title: '会员昵称',
      dataIndex: 'memberName',
      hideInTable: true,
    },
    {
      title: '评分',
      dataIndex: 'star',
      width: 100,
      hideInSearch: true,
      render: (_, record) => renderStarRating(record.star),
    },
    {
      title: '评价内容',
      dataIndex: 'content',
      width: 240,
      hideInSearch: true,
      ellipsis: true,
      render: (_, record) => (
        <Paragraph ellipsis={{ rows: 2, expandable: true, symbol: '展开' }} style={{ marginBottom: 0 }}>
          {record.content}
        </Paragraph>
      ),
    },
    {
      title: '审核状态',
      dataIndex: 'auditStatus',
      width: 120,
      valueType: 'select',
      valueEnum: auditStatusValueEnum,
      render: (_, record) => renderAuditStatusTag(record.auditStatus),
    },
    {
      title: '屏蔽状态',
      dataIndex: 'hidden',
      width: 110,
      valueType: 'select',
      valueEnum: hiddenValueEnum,
      render: (_, record) => renderHiddenTag(record.hidden),
    },
    {
      title: '审核人',
      dataIndex: 'auditorName',
      width: 120,
      hideInSearch: true,
      render: (_, record) => record.auditorName || '-',
    },
    {
      title: '审核时间',
      dataIndex: 'auditedAt',
      width: 160,
      hideInSearch: true,
      render: (_, record) => record.auditedAt || '-',
    },
    {
      title: '评价时间',
      dataIndex: 'createTime',
      width: 160,
      hideInSearch: true,
    },
    {
      title: '评价时间',
      dataIndex: 'createdAtRange',
      hideInTable: true,
      valueType: 'dateTimeRange',
      search: {
        transform: (value: any[]) => ({
          startTime: value?.[0]?.format('YYYY-MM-DD HH:mm:ss'),
          endTime: value?.[1]?.format('YYYY-MM-DD HH:mm:ss'),
        }),
      },
    },
    {
      title: '操作',
      valueType: 'option',
      fixed: 'right',
      width: 240,
      render: (_, record) => {
        const actions = [
          <Button key="detail" type="text" size="small" icon={<EyeOutlined />} onClick={() => openDetail(record.id)}>
            详情
          </Button>,
        ];

        if (record.appealStatus === 1) {
          actions.push(
            <Button
              key="appeal"
              type="text"
              size="small"
              onClick={() => openAppealModal(record)}
            >
              处理申诉
            </Button>,
          );
        }

        if (record.hidden === 1 || record.auditStatus === 3) {
          actions.push(
            <Button
              key="restore"
              type="text"
              size="small"
              icon={<RollbackOutlined />}
              onClick={() => confirmRestore(record)}
            >
              恢复
            </Button>,
          );
          return <Space>{actions}</Space>;
        }

        if (record.auditStatus !== 1) {
          actions.push(
            <Button
              key="approve"
              type="text"
              size="small"
              icon={<CheckCircleOutlined />}
              onClick={() => openAuditModal(record, 1)}
            >
              通过
            </Button>,
          );
        }

        actions.push(
          <Button
            key="reject"
            type="text"
            size="small"
            danger
            icon={<CloseCircleOutlined />}
            onClick={() => openAuditModal(record, 2)}
          >
            拒绝
          </Button>,
        );

        actions.push(
          <Button
            key="hide"
            type="text"
            size="small"
            icon={<StopOutlined />}
            onClick={() => openAuditModal(record, 3)}
          >
            屏蔽
          </Button>,
        );

        return <Space>{actions}</Space>;
      },
    },
  ];

  return (
    <PageContainer>
      <GovernanceScopeBar value={scope} onChange={setScope} entityLabel="商品评价治理" />
      <ProTable<CommentListItem>
        headerTitle="商品评价治理"
        actionRef={actionRef}
        rowKey="id"
        scroll={{ x: 1600 }}
        search={{ labelWidth: 'auto' }}
        rowSelection={{
          selectedRowKeys,
          onChange: (keys) => setSelectedRowKeys(keys),
        }}
        toolBarRender={() => [
          <Button key="batchApprove" type="primary" onClick={() => handleBatchAction(1, '批量通过')}>
            批量通过
          </Button>,
          <Button key="batchHide" danger onClick={() => handleBatchAction(0, '批量屏蔽')}>
            批量屏蔽
          </Button>,
        ]}
        request={async (params) => {
          const res = await queryCommentList(
            {
              current: params.current,
              pageSize: params.pageSize,
              productId: params.productId as number | undefined,
              productName: params.productName as string | undefined,
              memberName: params.memberName as string | undefined,
              showStatus:
                params.showStatus !== undefined && params.showStatus !== null
                  ? Number(params.showStatus)
                  : undefined,
              auditStatus:
                params.auditStatus !== undefined && params.auditStatus !== null
                  ? Number(params.auditStatus)
                  : undefined,
              hidden:
                params.hidden !== undefined && params.hidden !== null ? Number(params.hidden) : undefined,
              startTime: params.startTime as string | undefined,
              endTime: params.endTime as string | undefined,
            },
            scopePayload,
          );

          return {
            data: res.data || [],
            success: res.code === '000000',
            total: res.total || 0,
          };
        }}
        columns={columns}
      />

      <DrawerView
        open={drawerVisible}
        loading={detailLoading}
        comment={selectedComment}
        onClose={() => {
          setDrawerVisible(false);
          setSelectedComment(undefined);
        }}
        onApprove={() => selectedComment && openAuditModal(selectedComment, 1)}
        onReject={() => selectedComment && openAuditModal(selectedComment, 2)}
        onHide={() => selectedComment && openAuditModal(selectedComment, 3)}
        onRestore={() => selectedComment && confirmRestore(selectedComment)}
        onHandleAppeal={() => selectedComment && openAppealModal(selectedComment)}
      />

      <Modal
        title={
          auditModal.auditStatus === 1
            ? '审核通过评价'
            : auditModal.auditStatus === 2
              ? '审核拒绝评价'
              : '屏蔽评价'
        }
        open={auditModal.open}
        okText="提交"
        cancelText="取消"
        onCancel={() => {
          setAuditModal({ open: false, auditStatus: 1 });
          auditForm.resetFields();
        }}
        onOk={submitAudit}
      >
        <Form form={auditForm} layout="vertical">
          <Form.Item label="评价内容">
            <Paragraph style={{ marginBottom: 0 }}>{auditModal.record?.content || '-'}</Paragraph>
          </Form.Item>
          <Form.Item
            label={auditModal.auditStatus === 2 ? '拒绝原因' : '审核备注'}
            name="auditRemark"
            rules={
              auditModal.auditStatus === 2
                ? [{ required: true, message: '请填写拒绝原因' }]
                : undefined
            }
          >
            <Input.TextArea rows={4} maxLength={200} placeholder="请输入处理说明" />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="处理评价申诉"
        open={appealModal.open}
        okText="提交"
        cancelText="取消"
        onCancel={() => {
          setAppealModal({ open: false, appealStatus: 2 });
          appealForm.resetFields();
        }}
        onOk={submitAppeal}
      >
        <Form form={appealForm} layout="vertical">
          <Form.Item label="评价内容">
            <Paragraph style={{ marginBottom: 0 }}>{appealModal.record?.content || '-'}</Paragraph>
          </Form.Item>
          <Form.Item label="处理结果">
            <Radio.Group
              value={appealModal.appealStatus}
              onChange={(event) =>
                setAppealModal((current) => ({ ...current, appealStatus: event.target.value }))
              }
            >
              <Radio value={2}>申诉通过</Radio>
              <Radio value={3}>申诉驳回</Radio>
            </Radio.Group>
          </Form.Item>
          <Form.Item
            label="处理回复"
            name="appealReply"
            rules={[{ required: true, message: '请填写申诉处理回复' }]}
          >
            <Input.TextArea rows={4} maxLength={200} placeholder="请输入处理回复" />
          </Form.Item>
        </Form>
      </Modal>
    </PageContainer>
  );
};

type DrawerViewProps = {
  open: boolean;
  loading: boolean;
  comment?: DetailState;
  onClose: () => void;
  onApprove: () => void;
  onReject: () => void;
  onHide: () => void;
  onRestore: () => void;
  onHandleAppeal: () => void;
};

function DrawerView({
  open,
  loading,
  comment,
  onClose,
  onApprove,
  onReject,
  onHide,
  onRestore,
  onHandleAppeal,
}: DrawerViewProps) {
  const extra = !comment ? null : (
    <Space>
      {comment.appealStatus === 1 && <Button onClick={onHandleAppeal}>处理申诉</Button>}
      {(comment.hidden === 1 || comment.auditStatus === 3) && (
        <Button icon={<RollbackOutlined />} onClick={onRestore}>
          恢复评价
        </Button>
      )}
      {comment.hidden !== 1 && comment.auditStatus !== 3 && (
        <>
          {comment.auditStatus !== 1 && (
            <Button type="primary" icon={<CheckCircleOutlined />} onClick={onApprove}>
              审核通过
            </Button>
          )}
          <Button danger icon={<CloseCircleOutlined />} onClick={onReject}>
            审核拒绝
          </Button>
          <Button icon={<StopOutlined />} onClick={onHide}>
            屏蔽评价
          </Button>
        </>
      )}
    </Space>
  );

  return (
    <Drawer title="评价详情" width={760} open={open} onClose={onClose} extra={extra}>
      {loading ? (
        <div style={{ padding: 32, textAlign: 'center' }}>加载中...</div>
      ) : !comment ? (
        <div style={{ padding: 32, textAlign: 'center' }}>暂无数据</div>
      ) : (
        <>
          <Descriptions column={1} bordered size="small">
            <Descriptions.Item label="评价ID">{comment.id}</Descriptions.Item>
            <Descriptions.Item label="商品名称">
              <Space direction="vertical" size={2}>
                <Text>{comment.productName || comment.productAttribute || '商品评价'}</Text>
                <Text type="secondary">商品ID: {comment.productId}</Text>
              </Space>
            </Descriptions.Item>
            <Descriptions.Item label="会员信息">
              <Space>
                <Text>{comment.memberNickName || '-'}</Text>
                <Text type="secondary">会员ID: {comment.memberId}</Text>
              </Space>
            </Descriptions.Item>
            <Descriptions.Item label="评分">{renderStarRating(comment.star)}</Descriptions.Item>
            <Descriptions.Item label="评价内容">
              <Paragraph style={{ marginBottom: 0 }}>{comment.content || '-'}</Paragraph>
            </Descriptions.Item>
            <Descriptions.Item label="审核状态">{renderAuditStatusTag(comment.auditStatus)}</Descriptions.Item>
            <Descriptions.Item label="屏蔽状态">{renderHiddenTag(comment.hidden)}</Descriptions.Item>
            <Descriptions.Item label="审核人">{comment.auditorName || '-'}</Descriptions.Item>
            <Descriptions.Item label="审核时间">{comment.auditedAt || '-'}</Descriptions.Item>
            <Descriptions.Item label="审核备注">{comment.auditRemark || '-'}</Descriptions.Item>
            <Descriptions.Item label="申诉状态">
              {renderAppealStatusTag(comment.appealStatus)}
            </Descriptions.Item>
            <Descriptions.Item label="申诉原因">{comment.appealReason || '-'}</Descriptions.Item>
            <Descriptions.Item label="申诉回复">{comment.appealReply || '-'}</Descriptions.Item>
            <Descriptions.Item label="申诉时间">{comment.appealedAt || '-'}</Descriptions.Item>
            <Descriptions.Item label="申诉处理时间">{comment.appealHandledAt || '-'}</Descriptions.Item>
            <Descriptions.Item label="评价时间">{comment.createTime || '-'}</Descriptions.Item>
            <Descriptions.Item label="评价IP">{comment.memberIp || '-'}</Descriptions.Item>
            <Descriptions.Item label="点赞/阅读">
              {comment.collectCount} / {comment.readCount}
            </Descriptions.Item>
            <Descriptions.Item label="回复数">{comment.replayCount}</Descriptions.Item>
          </Descriptions>

          <Divider>晒图</Divider>
          {comment.pics ? (
            <Image.PreviewGroup>
              <Space wrap>
                {comment.pics
                  .split(',')
                  .filter((item) => item.trim())
                  .map((item) => (
                    <Image
                      key={item}
                      src={item}
                      width={88}
                      height={88}
                      style={{ objectFit: 'cover', borderRadius: 6 }}
                    />
                  ))}
              </Space>
            </Image.PreviewGroup>
          ) : (
            <Text type="secondary">暂无图片</Text>
          )}

          <Divider>回复记录</Divider>
          {comment.replays && comment.replays.length > 0 ? (
            <Space direction="vertical" size={12} style={{ width: '100%' }}>
              {comment.replays.map((item) => (
                <div
                  key={item.id}
                  style={{
                    border: '1px solid #F0F0F0',
                    borderRadius: 8,
                    padding: '12px 14px',
                    background: item.type === 1 ? '#F6FFED' : '#FAFAFA',
                  }}
                >
                  <Space size={8}>
                    <Tag color={item.type === 1 ? 'green' : 'blue'}>
                      {item.type === 1 ? '管理员' : '会员'}
                    </Tag>
                    <Text strong>{item.memberNickName || '-'}</Text>
                    <Text type="secondary">{item.createTime || '-'}</Text>
                  </Space>
                  <Paragraph style={{ margin: '8px 0 0' }}>{item.content || '-'}</Paragraph>
                </div>
              ))}
            </Space>
          ) : (
            <Text type="secondary">暂无回复记录</Text>
          )}

          <Divider>审核历史</Divider>
          {comment.auditLogs && comment.auditLogs.length > 0 ? (
            <Timeline>
              {comment.auditLogs.map((item) => (
                <Timeline.Item key={item.id}>
	                  <Space direction="vertical" size={2}>
	                    <Space size={8}>
	                      <Tag>{item.action}</Tag>
	                      <Text strong>{item.operatorName || '-'}</Text>
	                      <Text type="secondary">{item.createdAt || '-'}</Text>
	                    </Space>
	                    <Text type="secondary">
	                      状态变更: {item.fromStatus} {'->'} {item.toStatus}
	                    </Text>
	                    {item.remark ? <Paragraph style={{ marginBottom: 0 }}>{item.remark}</Paragraph> : null}
	                  </Space>
	                </Timeline.Item>
              ))}
            </Timeline>
          ) : (
            <Text type="secondary">暂无审核历史</Text>
          )}
        </>
      )}
    </Drawer>
  );
}

export default ProductCommentPage;
