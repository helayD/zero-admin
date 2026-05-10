import {
  DeleteOutlined,
  EditOutlined,
  ExclamationCircleOutlined,
  MoreOutlined,
  PlusOutlined,
} from '@ant-design/icons';
import {
  Alert,
  Button,
  Divider,
  Drawer,
  Dropdown,
  Menu,
  message,
  Modal,
  Select,
  Space,
  Tag,
  Tooltip,
  Typography,
} from 'antd';
import React, { useRef, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import type { ActionType, ProColumns } from '@ant-design/pro-table';
import ProTable from '@ant-design/pro-table';
import type { ProDescriptionsItemProps } from '@ant-design/pro-descriptions';
import ProDescriptions from '@ant-design/pro-descriptions';
import AddModal from './components/AddModal';
import BatchSetFulfillmentRuleModal from './components/BatchSetFulfillmentRuleModal';
import StatusActionModal, {
  type ProductSpuStatusModalAction,
} from './components/StatusActionModal';
import UpdateModal from './components/UpdateModal';
import type {
  ProductSpuDetailPayload,
  ProductSpuListItem,
  ProductSpuSubmitPayload,
} from './data.d';
import {
  addProductSpu,
  queryProductSpuDetail,
  queryProductSpuList,
  removeProductSpu,
  updateProductSpu,
  updateProductSpuStatus,
} from './service';
import { queryProductFulfillmentRuleList } from '@/pages/sms/ProductFulfillmentRule/service';
import SkuModal from '@/pages/pms/ProductSpu/components/SkuModal';
import { buildCatalogActionError, type CatalogActionError } from '@/pages/pms/errorFeedback';
import GovernanceScopeBar from '@/pages/system/components/GovernanceScopeBar';
import {
  buildGovernanceScopeLabel,
  defaultGovernanceScope,
  governanceScopeColor,
  type GovernanceScopeValue,
  toGovernancePayload,
} from '@/pages/system/components/governance';

const { confirm } = Modal;
const { Text } = Typography;

type ProductSpuStatusAction = 'publish' | 'verify' | 'recommend' | 'new' | 'delete';
type PendingStatusAction = {
  action: ProductSpuStatusModalAction;
  ids: number[];
  initialStatus?: number;
  currentProduct?: ProductSpuListItem;
};

const productSpuActionLabel: Record<ProductSpuStatusAction, string> = {
  publish: '上架状态',
  verify: '审核状态',
  recommend: '推荐状态',
  new: '新品状态',
  delete: '删除状态',
};

const renderPublishStatusTag = (status: number) =>
  status === 1 ? <Tag color="blue">已上架</Tag> : <Tag>已下架</Tag>;

const renderVerifyStatusTag = (status: number) => {
  switch (status) {
    case 1:
      return <Tag color="green">审核通过</Tag>;
    case 2:
      return <Tag color="red">审核驳回</Tag>;
    default:
      return <Tag>未审核</Tag>;
  }
};

const renderRecommendStatusTag = (status: number) =>
  status === 1 ? <Tag color="gold">推荐中</Tag> : <Tag>未推荐</Tag>;

const renderFulfillmentModeTag = (mode?: string) => {
  if (mode === 'digital_asset') {
    return <Tag color="purple">提货卡</Tag>;
  }
  if (mode === 'physical_delivery') {
    return <Tag color="blue">实物发货</Tag>;
  }
  return <Tag>未配置</Tag>;
};

const renderScopeSource = (record: Pick<
  ProductSpuListItem,
  'scopeType' | 'platformId' | 'tenantId' | 'merchantId'
>) => {
  const scopeLabel = buildGovernanceScopeLabel({
    scopeType: record.scopeType,
    platformId: record.platformId,
    tenantId: record.tenantId,
    merchantId: record.merchantId,
  });

  return <Tag color={governanceScopeColor(record.scopeType)}>{scopeLabel}</Tag>;
};

const renderReviewSummary = (record: Pick<
  ProductSpuListItem,
  'verifyStatus' | 'reviewMan' | 'reviewTime' | 'reviewDetail'
>) => {
  if (!record.reviewMan && !record.reviewTime && !record.reviewDetail) {
    return <Text type="secondary">暂无审核记录</Text>;
  }

  return (
    <Space direction="vertical" size={2}>
      <Space size={6} wrap>
        {renderVerifyStatusTag(record.verifyStatus)}
        {record.reviewMan ? <Text>{record.reviewMan}</Text> : null}
        {record.reviewTime ? <Text type="secondary">{record.reviewTime}</Text> : null}
      </Space>
      {record.reviewDetail ? (
        <Text type={record.verifyStatus === 2 ? 'danger' : undefined}>{record.reviewDetail}</Text>
      ) : (
        <Text type="secondary">本次审核未填写说明</Text>
      )}
    </Space>
  );
};

const renderPublishSummary = (record: Pick<
  ProductSpuListItem,
  'publishStatus' | 'publishMan' | 'publishTime' | 'publishDetail'
>) => {
  if (!record.publishMan && !record.publishTime && !record.publishDetail) {
    return <Text type="secondary">暂无上下架记录</Text>;
  }

  return (
    <Space direction="vertical" size={2}>
      <Space size={6} wrap>
        {renderPublishStatusTag(record.publishStatus)}
        {record.publishMan ? <Text>{record.publishMan}</Text> : null}
        {record.publishTime ? <Text type="secondary">{record.publishTime}</Text> : null}
      </Space>
      {record.publishDetail ? (
        <Text type={record.publishStatus === 0 ? 'danger' : undefined}>{record.publishDetail}</Text>
      ) : (
        <Text type="secondary">本次上下架未填写说明</Text>
      )}
    </Space>
  );
};

const renderRecommendSummary = (record: Pick<
  ProductSpuListItem,
  'recommendStatus' | 'recommendMan' | 'recommendTime' | 'recommendDetail'
>) => {
  if (!record.recommendMan && !record.recommendTime && !record.recommendDetail) {
    return <Text type="secondary">暂无推荐记录</Text>;
  }

  return (
    <Space direction="vertical" size={2}>
      <Space size={6} wrap>
        {renderRecommendStatusTag(record.recommendStatus)}
        {record.recommendMan ? <Text>{record.recommendMan}</Text> : null}
        {record.recommendTime ? <Text type="secondary">{record.recommendTime}</Text> : null}
      </Space>
      {record.recommendDetail ? (
        <Text>{record.recommendDetail}</Text>
      ) : (
        <Text type="secondary">本次推荐调整未填写说明</Text>
      )}
    </Space>
  );
};

/**
 * 添加商品SPU
 * @param fields
 */
const handleAdd = async (
  fields: ProductSpuSubmitPayload,
  scope: GovernanceScopeValue,
  onError?: (error: CatalogActionError) => void,
) => {
  const hide = message.loading('正在添加');
  try {
    await addProductSpu({ ...fields, ...toGovernancePayload(scope) });
    hide();
    message.success('添加成功');
    return true;
  } catch (error) {
    hide();
    const catalogError = buildCatalogActionError(error, '商品 SPU 建档失败');
    onError?.(catalogError);
    message.error(catalogError.description);
    return false;
  }
};

/**
 * 更新商品SPU
 * @param fields
 */
const handleUpdate = async (
  fields: ProductSpuSubmitPayload & GovernanceScopeValue,
  onError?: (error: CatalogActionError) => void,
) => {
  const hide = message.loading('正在更新');
  try {
    await updateProductSpu(fields);
    hide();

    message.success('更新成功');
    return true;
  } catch (error) {
    hide();
    const catalogError = buildCatalogActionError(error, '商品 SPU 更新失败');
    onError?.(catalogError);
    message.error(catalogError.description);
    return false;
  }
};

const buildSubmitPayloadFromDetail = (
  detail?: ProductSpuDetailPayload,
): ProductSpuSubmitPayload | undefined => {
  if (!detail?.productData) {
    return undefined;
  }

  return {
    ...(detail.productData as ProductSpuSubmitPayload),
    ladderList: Array.isArray(detail.ladderList) ? detail.ladderList : [],
    fullList: Array.isArray(detail.fullList) ? detail.fullList : [],
    memberPriceList: Array.isArray(detail.memberPriceList) ? detail.memberPriceList : [],
    skuList: Array.isArray(detail.skuList) ? detail.skuList : [],
    attributeValueList: Array.isArray(detail.attributeValueList) ? detail.attributeValueList : [],
    subjectIds: Array.isArray(detail.subjectIds) ? detail.subjectIds : [],
    prefrenceAreaIds: Array.isArray(detail.prefrenceAreaIds) ? detail.prefrenceAreaIds : [],
  };
};

/**
 *  删除商品SPU
 * @param ids
 */
const handleRemove = async (ids: number[], scope: GovernanceScopeValue) => {
  const hide = message.loading('正在删除');
  if (ids.length === 0) return true;
  try {
    await removeProductSpu(ids, toGovernancePayload(scope));
    hide();
    message.success('删除成功，即将刷新');
    return true;
  } catch (error) {
    hide();
    return false;
  }
};

/**
 * 更新商品SPU状态
 * @param action
 * @param ids
 * @param status
 */
const handleStatus = async (
  action: ProductSpuStatusAction,
  ids: number[],
  status: number,
  scope: GovernanceScopeValue,
  detail?: string,
  onError?: (error: CatalogActionError) => void,
) => {
  const hide = message.loading('正在更新状态');
  if (ids.length == 0) {
    hide();
    return true;
  }
  try {
    await updateProductSpuStatus(action, {
      ids,
      status,
      detail,
      ...toGovernancePayload(scope),
    });
    hide();
    message.success(`${productSpuActionLabel[action]}更新成功`);
    return true;
  } catch (error) {
    hide();
    const catalogError = buildCatalogActionError(error, `${productSpuActionLabel[action]}更新失败`);
    onError?.(catalogError);
    message.error(catalogError.description);
    return false;
  }
};

const ProductSpuList: React.FC = () => {
  const [addVisible, handleAddVisible] = useState<boolean>(false);
  const [updateVisible, handleUpdateVisible] = useState<boolean>(false);
  const [showDetail, setShowDetail] = useState<boolean>(false);
  const actionRef = useRef<ActionType>();
  const [currentRow, setCurrentRow] = useState<ProductSpuListItem>();
  const [currentSubmitData, setCurrentSubmitData] = useState<ProductSpuSubmitPayload>();
  const [skuVisible, handleSkuVisible] = useState<boolean>(false);
  const [scope, setScope] = useState<GovernanceScopeValue>(defaultGovernanceScope);
  const [submitError, setSubmitError] = useState<CatalogActionError>();
  const [pendingStatusAction, setPendingStatusAction] = useState<PendingStatusAction>();
  const [batchRuleModalVisible, setBatchRuleModalVisible] = useState(false);
  const [batchRuleTargets, setBatchRuleTargets] = useState<ProductSpuListItem[]>([]);
  const [quickView, setQuickView] = useState<
    'all' | 'pendingReview' | 'onShelf' | 'offShelf' | 'recommended' | 'digitalAsset' | 'physical'
  >('all');
  // A3: ruleMap 存完整 rule 对象（不只名称），用于 Tooltip 展示关键属性
  type RuleBrief = {
    ruleName: string;
    cardTemplateName?: string;
    expireDays?: number;
    transferable?: number;
    transferLimit?: number;
    ruleStatus?: number;
  };
  const [ruleMap, setRuleMap] = useState<Record<number, RuleBrief>>({});

  const quickViewParams =
    quickView === 'pendingReview'
      ? { verifyStatus: 0 }
      : quickView === 'onShelf'
        ? { publishStatus: 1 }
        : quickView === 'offShelf'
          ? { publishStatus: 0 }
          : quickView === 'recommended'
            ? { recommendStatus: 1 }
            : quickView === 'digitalAsset'
              ? { fulfillmentMode: 'digital_asset' }
              : quickView === 'physical'
                ? { fulfillmentMode: 'physical_delivery' }
                : {};

  React.useEffect(() => {
    queryProductFulfillmentRuleList({
      pageSize: 999,
      ...toGovernancePayload(scope),
    })
      .then((res) => {
        const map: Record<number, RuleBrief> = {};
        (res.data || []).forEach((item: any) => {
          map[item.id] = {
            ruleName: item.ruleName,
            cardTemplateName: item.cardTemplateName,
            expireDays: item.expireDays,
            transferable: item.transferable,
            transferLimit: item.transferLimit,
            ruleStatus: item.ruleStatus,
          };
        });
        setRuleMap(map);
      })
      .catch(() => {
        setRuleMap({});
      });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [scope]);

  const openUpdateModal = async (record: ProductSpuListItem) => {
    setSubmitError(undefined);
    const hide = message.loading('正在加载商品详情...');
    try {
      const detailResp = await queryProductSpuDetail(record.id, toGovernancePayload(scope));
      const submitDraft = buildSubmitPayloadFromDetail(detailResp.data);
      setCurrentRow({ ...record, ...(detailResp.data?.productData || {}) });
      setCurrentSubmitData(submitDraft);
      handleUpdateVisible(true);
    } catch (error) {
      const catalogError = buildCatalogActionError(error, '商品 SPU 详情加载失败');
      setSubmitError(catalogError);
      message.error(catalogError.description);
    } finally {
      hide();
    }
  };

  const showDeleteConfirm = (ids: number[]) => {
    confirm({
      title: '是否删除记录?',
      icon: <ExclamationCircleOutlined />,
      content: `当前主体：${buildGovernanceScopeLabel(scope)}。删除后不可恢复，请确认。`,
      onOk() {
        handleRemove(ids, scope).then(() => {
          actionRef.current?.reloadAndRest?.();
        });
      },
      onCancel() {},
    });
  };

  const openStatusActionModal = (
    action: ProductSpuStatusModalAction,
    ids: number[],
    initialStatus?: number,
    currentProduct?: ProductSpuListItem,
  ) => {
    setSubmitError(undefined);
    setPendingStatusAction({ action, ids, initialStatus, currentProduct });
  };

  const columns: ProColumns<ProductSpuListItem>[] = [
    {
      title: '编号',
      dataIndex: 'id',
      hideInSearch: true,
      width: 70,
    },
    {
      title: '商品名称',
      dataIndex: 'name',
      width: 240,
      ellipsis: true,
      render: (dom, entity) => {
        return (
          <a
            onClick={() => {
              setCurrentRow(entity);
              setShowDetail(true);
            }}
          >
            {dom}
          </a>
        );
      },
    },
    {
      title: '商品货号',
      dataIndex: 'productSn',
      width: 110,
    },

    {
      title: '商品分类ID',
      dataIndex: 'categoryId',
      hideInSearch: true,
      hideInTable: true,
    },
    {
      title: '商品分类ID集合',
      dataIndex: 'categoryIds',
      hideInSearch: true,
      hideInTable: true,
    },
    {
      title: '商品分类名称',
      dataIndex: 'categoryName',
      hideInSearch: true,
      hideInTable: true,
      render: (dom, entity) => {
        return (
          <a
            onClick={() => {
              setCurrentRow(entity);
              setShowDetail(true);
            }}
          >
            {dom}
          </a>
        );
      },
    },

    {
      title: '品牌ID',
      dataIndex: 'brandId',
      hideInSearch: true,
      hideInTable: true,
    },
    {
      title: '品牌名称',
      dataIndex: 'brandName',
      hideInSearch: true,
      hideInTable: true,
      render: (dom, entity) => {
        return (
          <a
            onClick={() => {
              setCurrentRow(entity);
              setShowDetail(true);
            }}
          >
            {dom}
          </a>
        );
      },
    },

    {
      title: '单位',
      dataIndex: 'unit',
      hideInSearch: true,
      hideInTable: true,
    },
    {
      title: '重量(kg)',
      dataIndex: 'weight',
      hideInSearch: true,
      hideInTable: true,
    },
    {
      title: '关键词',
      dataIndex: 'keywords',
      hideInSearch: true,
      hideInTable: true,
    },
    {
      title: '简介',
      dataIndex: 'brief',
      hideInSearch: true,
      hideInTable: true,
    },
    {
      title: '详细描述',
      dataIndex: 'description',
      hideInSearch: true,
      hideInTable: true,
    },
    {
      title: '画册图片，最多8张，以逗号分割',
      dataIndex: 'albumPics',
      hideInSearch: true,
      hideInTable: true,
    },
    {
      title: '主图',
      dataIndex: 'mainPic',
      hideInSearch: true,
      valueType: 'image',
      width: 100,
      fieldProps: { width: 100, height: 80 },
    },
    {
      title: '价格区间',
      dataIndex: 'priceRange',
      hideInSearch: true,
      width: 130,
    },
    {
      title: '作用域',
      dataIndex: 'scopeType',
      hideInSearch: true,
      width: 110,
      render: (_, entity) => renderScopeSource(entity),
    },
    {
      title: '上架状态',
      dataIndex: 'publishStatus',
      width: 90,
      renderFormItem: (text, row) => {
        return (
          <Select
            value={row.value}
            options={[
              { value: 1, label: '上架' },
              { value: 0, label: '下架' },
            ]}
          />
        );
      },
      render: (_, entity) => renderPublishStatusTag(entity.publishStatus),
    },
    {
      title: '最近上下架说明',
      dataIndex: 'publishDetail',
      hideInSearch: true,
      hideInTable: true,
      width: 260,
      render: (_, entity) => renderPublishSummary(entity),
    },
    {
      title: '是否新品',
      dataIndex: 'newStatus',
      hideInTable: true,
      renderFormItem: (text, row) => {
        return (
          <Select
            value={row.value}
            options={[
              { value: 1, label: '是' },
              { value: 0, label: '不是' },
            ]}
          />
        );
      },
      render: (_, entity) =>
        entity.newStatus === 1 ? <Tag color="cyan">新品</Tag> : <Tag>不是</Tag>,
    },
    {
      title: '是否推荐',
      dataIndex: 'recommendStatus',
      width: 90,
      renderFormItem: (text, row) => {
        return (
          <Select
            value={row.value}
            options={[
              { value: 1, label: '推荐' },
              { value: 0, label: '不推荐' },
            ]}
          />
        );
      },
      render: (_, entity) => renderRecommendStatusTag(entity.recommendStatus),
    },
    {
      title: '最近推荐反馈',
      dataIndex: 'recommendDetail',
      hideInSearch: true,
      hideInTable: true,
      width: 260,
      render: (_, entity) => renderRecommendSummary(entity),
    },
    {
      title: '审核状态',
      dataIndex: 'verifyStatus',
      width: 100,
      renderFormItem: (text, row) => {
        return (
          <Select
            value={row.value}
            options={[
              { value: 1, label: '审核通过' },
              { value: 0, label: '未审核' },
              { value: 2, label: '审核驳回' },
            ]}
          />
        );
      },
      render: (_, entity) => renderVerifyStatusTag(entity.verifyStatus),
    },
    {
      title: '最新审核反馈',
      dataIndex: 'reviewDetail',
      hideInSearch: true,
      hideInTable: true,
      width: 260,
      render: (_, entity) => renderReviewSummary(entity),
    },
    {
      title: '预告商品',
      dataIndex: 'previewStatus',
      hideInTable: true,
      renderFormItem: (text, row) => {
        return (
          <Select
            value={row.value}
            options={[
              { value: 1, label: '是' },
              { value: 0, label: '不是' },
            ]}
          />
        );
      },
      render: (dom, entity) => {
        return entity.previewStatus == 1 ? <Tag color="gold">预告中</Tag> : <Tag>普通</Tag>;
      },
    },
    {
      title: '排序',
      dataIndex: 'sort',
      hideInSearch: true,
      hideInTable: true,
    },
    {
      title: '新品排序',
      dataIndex: 'newStatusSort',
      hideInSearch: true,
      hideInTable: true,
    },
    {
      title: '推荐排序',
      dataIndex: 'recommendStatusSort',
      hideInSearch: true,
      hideInTable: true,
    },
    {
      title: '销量',
      dataIndex: 'sales',
      hideInSearch: true,
      hideInTable: true,
    },
    {
      title: '库存',
      dataIndex: 'stock',
      hideInSearch: true,
      width: 80,
    },
    {
      title: '预警库存',
      dataIndex: 'lowStock',
      hideInSearch: true,
      hideInTable: true,
    },
    {
      title: '促销类型',
      dataIndex: 'promotionType',
      width: 100,
      renderFormItem: (text, row) => {
        return (
          <Select
            value={row.value}
            options={[
              { value: 0, label: '原价' },
              { value: 1, label: '促销价' },
              { value: 2, label: '会员价' },
              { value: 3, label: '阶梯价格' },
              { value: 4, label: '满减价格' },
              { value: 5, label: '秒杀价格' },
            ]}
          />
        );
      },
      render: (dom, entity) => {
        switch (entity.promotionType) {
          case 0:
            return <Tag color={'success'}>使用原价</Tag>;
          case 1:
            return <Tag color={'success'}>使用促销价</Tag>;
          case 2:
            return <Tag color={'success'}>使用会员价</Tag>;
          case 3:
            return <Tag color={'success'}>使用阶梯价格</Tag>;
          case 4:
            return <Tag color={'success'}>使用满减价格</Tag>;
          case 5:
            return <Tag color={'success'}>秒杀</Tag>;
        }
        return <>未知{entity.promotionType}</>;
      },
    },
    {
      title: '履约模式',
      dataIndex: 'fulfillmentMode',
      width: 110,
      // Story 10.10 Task 7: 履约模式可作为搜索条件，服务端过滤
      valueType: 'select',
      valueEnum: {
        physical_delivery: { text: '实物发货' },
        digital_asset: { text: '提货卡' },
      },
      render: (dom, entity) => {
        return renderFulfillmentModeTag(entity.fulfillmentMode);
      },
    },
    {
      title: '发卡规则',
      dataIndex: 'fulfillmentRuleId',
      hideInSearch: true,
      width: 160,
      ellipsis: true,
      render: (_, entity) => {
        if (!entity.fulfillmentRuleId) return '-';
        const brief = ruleMap[entity.fulfillmentRuleId];
        if (!brief) {
          return <Text type="warning">规则ID: {entity.fulfillmentRuleId}（未加载/不在范围）</Text>;
        }
        // A3: Tooltip 展示规则关键属性
        const transferText =
          brief.transferable === 1
            ? brief.transferLimit && brief.transferLimit > 0
              ? `可转赠（限 ${brief.transferLimit} 次）`
              : '可转赠'
            : '不可转赠';
        return (
          <Tooltip
            title={
              <Space direction="vertical" size={2}>
                <span>规则名：{brief.ruleName}</span>
                {brief.cardTemplateName ? <span>卡片模板：{brief.cardTemplateName}</span> : null}
                {typeof brief.expireDays === 'number' ? (
                  <span>有效期：{brief.expireDays > 0 ? `${brief.expireDays} 天` : '永久'}</span>
                ) : null}
                <span>转赠：{transferText}</span>
                {brief.ruleStatus === 0 ? <span style={{ color: '#ff4d4f' }}>当前状态：已禁用</span> : null}
              </Space>
            }
          >
            <span style={{ borderBottom: '1px dashed #999', cursor: 'help' }}>{brief.ruleName}</span>
          </Tooltip>
        );
      },
    },

    {
      title: '详情标题',
      dataIndex: 'detailTitle',
      hideInSearch: true,
      hideInTable: true,
    },
    {
      title: '详情描述',
      dataIndex: 'detailDesc',
      hideInSearch: true,
      hideInTable: true,
    },
    {
      title: '产品详情网页内容',
      dataIndex: 'detailHtml',
      hideInSearch: true,
      hideInTable: true,
    },
    {
      title: '移动端网页详情',
      dataIndex: 'detailMobileHtml',
      hideInSearch: true,
      hideInTable: true,
    },
    {
      title: '创建人ID',
      dataIndex: 'createBy',
      hideInSearch: true,
      hideInTable: true,
    },
    {
      title: '创建时间',
      dataIndex: 'createTime',
      hideInSearch: true,
      hideInTable: true,
    },
    {
      title: '更新人ID',
      dataIndex: 'updateBy',
      hideInSearch: true,
      hideInTable: true,
    },
    {
      title: '更新时间',
      dataIndex: 'updateTime',
      hideInSearch: true,
      hideInTable: true,
    },
    {
      title: '操作',
      dataIndex: 'option',
      valueType: 'option',
      width: 240,
      fixed: 'right',
      render: (_, record) => {
        const moreMenu = (
          <Menu
            items={[
              {
                key: 'publish-toggle',
                label: record.publishStatus === 1 ? '下架' : '上架',
                onClick: () =>
                  openStatusActionModal(
                    'publish',
                    [record.id],
                    record.publishStatus === 1 ? 0 : 1,
                    record,
                  ),
              },
              {
                key: 'verify-pass',
                label: '审核通过',
                disabled: record.verifyStatus === 1,
                onClick: () => openStatusActionModal('verify', [record.id], 1, record),
              },
              {
                key: 'verify-reject',
                label: '审核驳回',
                disabled: record.verifyStatus === 2,
                onClick: () => openStatusActionModal('verify', [record.id], 2, record),
              },
              {
                key: 'recommend-toggle',
                label: record.recommendStatus === 1 ? '取消推荐' : '设为推荐',
                onClick: () =>
                  openStatusActionModal(
                    'recommend',
                    [record.id],
                    record.recommendStatus === 1 ? 0 : 1,
                    record,
                  ),
              },
              {
                key: 'new-toggle',
                label: record.newStatus === 1 ? '取消新品' : '设为新品',
                onClick: async () => {
                  const success = await handleStatus(
                    'new',
                    [record.id],
                    record.newStatus === 1 ? 0 : 1,
                    scope,
                    undefined,
                    setSubmitError,
                  );
                  if (success) {
                    actionRef.current?.reload?.();
                  }
                },
              },
            ]}
          />
        );
        return (
          <Space size={4} split={<Divider type="vertical" style={{ margin: 0 }} />}>
            <a
              key="edit"
              onClick={() => {
                void openUpdateModal(record);
              }}
            >
              <EditOutlined /> 编辑
            </a>
            <a
              key="sku"
              onClick={() => {
                handleSkuVisible(true);
                setCurrentRow(record);
              }}
            >
              规格
            </a>
            <Dropdown overlay={moreMenu} trigger={['click']}>
              <a key="more">
                <MoreOutlined /> 更多
              </a>
            </Dropdown>
            <a
              key="delete"
              style={{ color: '#ff4d4f' }}
              onClick={() => {
                showDeleteConfirm([record.id]);
              }}
            >
              <DeleteOutlined /> 删除
            </a>
          </Space>
        );
      },
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
        entityLabel="商品 SPU"
        style={{ marginBottom: 16 }}
      />
      <Alert
        showIcon
        type="info"
        style={{ marginBottom: 16 }}
        message={`当前查询范围：${buildGovernanceScopeLabel(scope)}`}
        description="平台管理员可切换到租户/商户视角查看商品；租户和商户账号只会看到自己的主体数据。"
      />
      {submitError && (
        <Alert
          showIcon
          closable
          type="error"
          style={{ marginBottom: 16 }}
          message={submitError.title}
          description={submitError.description}
          onClose={() => setSubmitError(undefined)}
        />
      )}
      <ProTable<ProductSpuListItem>
        headerTitle="商品SPU管理"
        actionRef={actionRef}
        rowKey="id"
        search={{
          labelWidth: 120,
        }}
        toolBarRender={() => [
          <Space key="quick-view" size={8} wrap>
            <Button
              type={quickView === 'all' ? 'primary' : 'default'}
              onClick={() => {
                setQuickView('all');
                actionRef.current?.reload?.();
              }}
            >
              全部
            </Button>
            <Button
              type={quickView === 'pendingReview' ? 'primary' : 'default'}
              onClick={() => {
                setQuickView('pendingReview');
                actionRef.current?.reload?.();
              }}
            >
              待审核
            </Button>
            <Button
              type={quickView === 'onShelf' ? 'primary' : 'default'}
              onClick={() => {
                setQuickView('onShelf');
                actionRef.current?.reload?.();
              }}
            >
              已上架
            </Button>
            <Button
              type={quickView === 'offShelf' ? 'primary' : 'default'}
              onClick={() => {
                setQuickView('offShelf');
                actionRef.current?.reload?.();
              }}
            >
              已下架
            </Button>
            <Button
              type={quickView === 'recommended' ? 'primary' : 'default'}
              onClick={() => {
                setQuickView('recommended');
                actionRef.current?.reload?.();
              }}
            >
              推荐中
            </Button>
            <Divider type="vertical" style={{ height: 24 }} />
            <Button
              type={quickView === 'digitalAsset' ? 'primary' : 'default'}
              onClick={() => {
                setQuickView('digitalAsset');
                actionRef.current?.reload?.();
              }}
            >
              提货卡商品
            </Button>
            <Button
              type={quickView === 'physical' ? 'primary' : 'default'}
              onClick={() => {
                setQuickView('physical');
                actionRef.current?.reload?.();
              }}
            >
              实物商品
            </Button>
          </Space>,
          <Button
            type="primary"
            key="primary"
            onClick={() => {
              setSubmitError(undefined);
              setCurrentRow(undefined);
              setCurrentSubmitData(undefined);
              handleAddVisible(true);
            }}
          >
            <PlusOutlined /> 新增
          </Button>,
        ]}
        request={(params) =>
          queryProductSpuList({
            ...params,
            ...quickViewParams,
            ...toGovernancePayload(scope),
          })
        }
        columns={columns}
        rowSelection={{}}
        pagination={{ pageSize: 10 }}
        scroll={{ x: 'max-content' }}
        tableAlertRender={({ selectedRowKeys, selectedRows }) => {
          const ids = selectedRows.map((row) => row.id);
          return (
            <Space size={16} wrap>
              <span>已选 {selectedRowKeys.length} 项</span>
              <Button
                icon={<EditOutlined />}
                style={{ borderRadius: '5px' }}
                onClick={async () => {
                  openStatusActionModal('publish', ids, 1);
                }}
              >
                批量上架
              </Button>
              <Button
                icon={<EditOutlined />}
                style={{ borderRadius: '5px' }}
                onClick={async () => {
                  openStatusActionModal('publish', ids, 0);
                }}
              >
                批量下架
              </Button>
              <Button
                icon={<EditOutlined />}
                style={{ borderRadius: '5px' }}
                onClick={async () => {
                  openStatusActionModal('verify', ids, 1);
                }}
              >
                批量审核通过
              </Button>
              <Button
                icon={<EditOutlined />}
                style={{ borderRadius: '5px' }}
                onClick={async () => {
                  openStatusActionModal('verify', ids, 2);
                }}
              >
                批量驳回
              </Button>
              <Button
                icon={<EditOutlined />}
                style={{ borderRadius: '5px' }}
                onClick={() => {
                  setBatchRuleTargets(selectedRows);
                  setBatchRuleModalVisible(true);
                }}
              >
                批量设置发卡规则
              </Button>
              <Button
                icon={<DeleteOutlined />}
                danger
                style={{ borderRadius: '5px' }}
                onClick={async () => {
                  showDeleteConfirm(ids);
                }}
              >
                批量删除
              </Button>
            </Space>
          );
        }}
      />

      <StatusActionModal
        visible={!!pendingStatusAction}
        action={pendingStatusAction?.action}
        scopeLabel={buildGovernanceScopeLabel(scope)}
        targetCount={pendingStatusAction?.ids.length || 0}
        initialStatus={pendingStatusAction?.initialStatus}
        currentProduct={pendingStatusAction?.currentProduct}
        onCancel={() => setPendingStatusAction(undefined)}
        onSubmit={async ({ status, detail }) => {
          if (!pendingStatusAction) {
            return;
          }
          const success = await handleStatus(
            pendingStatusAction.action,
            pendingStatusAction.ids,
            status,
            scope,
            detail,
            setSubmitError,
          );
          if (success) {
            setPendingStatusAction(undefined);
            actionRef.current?.clearSelected?.();
            actionRef.current?.reload?.();
          }
        }}
      />

      <AddModal
        key={'AddModal'}
        onSubmit={async (value) => {
          setSubmitError(undefined);
          const success = await handleAdd(value, scope, setSubmitError);
          if (success) {
            handleAddVisible(false);
            setCurrentRow(undefined);
            setCurrentSubmitData(undefined);
            if (actionRef.current) {
              actionRef.current.reload();
            }
          }
        }}
        onCancel={() => {
          handleAddVisible(false);
          setCurrentSubmitData(undefined);
          if (!showDetail) {
            setCurrentRow(undefined);
          }
        }}
        addVisible={addVisible}
        scope={scope}
        submitError={submitError}
      />

      <UpdateModal
        key={'UpdateModal'}
        onSubmit={async (value) => {
          setSubmitError(undefined);
          const success = await handleUpdate(
            { ...(currentSubmitData || {}), ...value, ...toGovernancePayload(scope) },
            setSubmitError,
          );
          if (success) {
            handleUpdateVisible(false);
            setCurrentRow(undefined);
            setCurrentSubmitData(undefined);
            if (actionRef.current) {
              actionRef.current.reload();
            }
          }
        }}
        onCancel={() => {
          handleUpdateVisible(false);
          setCurrentSubmitData(undefined);
          if (!showDetail) {
            setCurrentRow(undefined);
          }
        }}
        updateVisible={updateVisible}
        currentData={currentSubmitData || currentRow || {}}
        scope={scope}
        submitError={submitError}
      />
      <BatchSetFulfillmentRuleModal
        visible={batchRuleModalVisible}
        selectedRows={batchRuleTargets}
        scope={scope}
        onCancel={() => {
          setBatchRuleModalVisible(false);
          setBatchRuleTargets([]);
        }}
        onSuccess={() => {
          setBatchRuleModalVisible(false);
          setBatchRuleTargets([]);
          actionRef.current?.clearSelected?.();
          actionRef.current?.reload?.();
        }}
      />
      <SkuModal
        key={'SkuModal'}
        onCancel={() => {
          handleSkuVisible(false);
          if (!showDetail) {
            setCurrentRow(undefined);
          }
        }}
        modalVisible={skuVisible}
        spuId={currentRow?.id || 0}
        scope={scope}
      />
      <Drawer
        width={600}
        visible={showDetail}
        onClose={() => {
          setCurrentRow(undefined);
          setShowDetail(false);
        }}
        closable={false}
      >
        {currentRow?.id && (
          <ProDescriptions<ProductSpuListItem>
            column={2}
            title={'商品SPU详情'}
            request={async () => ({
              data: currentRow || {},
            })}
            params={{
              id: currentRow?.id,
            }}
            columns={columns as ProDescriptionsItemProps<ProductSpuListItem>[]}
          />
        )}
      </Drawer>
    </PageContainer>
  );
};

export default ProductSpuList;
