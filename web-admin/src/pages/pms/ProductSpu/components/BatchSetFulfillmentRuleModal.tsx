import React, { useEffect, useState } from 'react';
import { Alert, Modal, Progress, Select, Space, Typography } from 'antd';
import { queryProductFulfillmentRuleList } from '@/pages/sms/ProductFulfillmentRule/service';
import { queryProductSpuDetail, updateProductSpu } from '@/pages/pms/ProductSpu/service';
import type { GovernanceScopeValue } from '@/pages/system/components/governance';
import { toGovernancePayload } from '@/pages/system/components/governance';
import type { ProductSpuListItem, ProductSpuSubmitPayload } from '@/pages/pms/ProductSpu/data.d';

const { Text } = Typography;

type RuleOption = {
  label: string;
  value: number;
  ruleName: string;
  cardTemplateName?: string;
  expireDays?: number;
};

export interface BatchSetFulfillmentRuleModalProps {
  visible: boolean;
  onCancel: () => void;
  onSuccess: () => void;
  selectedRows: ProductSpuListItem[];
  scope: GovernanceScopeValue;
}

type RowResult = {
  id: number;
  name: string;
  ok: boolean;
  error?: string;
};

const BatchSetFulfillmentRuleModal: React.FC<BatchSetFulfillmentRuleModalProps> = ({
  visible,
  onCancel,
  onSuccess,
  selectedRows,
  scope,
}) => {
  const [ruleOptions, setRuleOptions] = useState<RuleOption[]>([]);
  const [selectedRuleId, setSelectedRuleId] = useState<number | undefined>(undefined);
  const [running, setRunning] = useState(false);
  const [progressDone, setProgressDone] = useState(0);
  const [results, setResults] = useState<RowResult[]>([]);

  useEffect(() => {
    if (!visible) {
      setSelectedRuleId(undefined);
      setProgressDone(0);
      setResults([]);
      return;
    }
    queryProductFulfillmentRuleList({
      pageSize: 999,
      ruleStatus: 1,
      ...toGovernancePayload(scope),
    })
      .then((res) => {
        const opts: RuleOption[] = (res.data || []).map((item: any) => ({
          label: `${item.ruleName}${item.cardTemplateName ? ` · ${item.cardTemplateName}` : ''}${
            typeof item.expireDays === 'number' && item.expireDays > 0
              ? ` · ${item.expireDays}天`
              : ''
          }`,
          value: item.id,
          ruleName: item.ruleName,
          cardTemplateName: item.cardTemplateName,
          expireDays: item.expireDays,
        }));
        setRuleOptions(opts);
      })
      .catch(() => {
        setRuleOptions([]);
      });
  }, [visible, scope]);

  const totalCount = selectedRows.length;

  const handleConfirm = async () => {
    if (!selectedRuleId) {
      return;
    }
    setRunning(true);
    setProgressDone(0);
    const acc: RowResult[] = [];
    const scopePayload = toGovernancePayload(scope);

    for (let i = 0; i < selectedRows.length; i++) {
      const row = selectedRows[i];
      try {
        const detailResp = await queryProductSpuDetail(row.id, scopePayload);
        const detail = detailResp?.data;
        if (!detail || !detail.productData) {
          throw new Error('商品详情为空');
        }
        // ProductSpuSubmitPayload 是扁平结构：productData 字段平铺到顶层 + nested 列表
        const payload: ProductSpuSubmitPayload = {
          ...(detail.productData || {}),
          fulfillmentMode: 'digital_asset',
          fulfillmentRuleId: selectedRuleId,
          ladderList: Array.isArray(detail.ladderList) ? detail.ladderList : [],
          fullList: Array.isArray(detail.fullList) ? detail.fullList : [],
          memberPriceList: Array.isArray(detail.memberPriceList) ? detail.memberPriceList : [],
          skuList: Array.isArray(detail.skuList) ? detail.skuList : [],
          attributeValueList: Array.isArray(detail.attributeValueList)
            ? detail.attributeValueList
            : [],
          subjectIds: Array.isArray(detail.subjectIds) ? detail.subjectIds : [],
          prefrenceAreaIds: Array.isArray(detail.prefrenceAreaIds) ? detail.prefrenceAreaIds : [],
          ...scopePayload,
        };
        await updateProductSpu(payload);
        acc.push({ id: row.id, name: row.name || `#${row.id}`, ok: true });
      } catch (error: any) {
        acc.push({
          id: row.id,
          name: row.name || `#${row.id}`,
          ok: false,
          error: error?.message || '未知错误',
        });
      }
      setProgressDone(i + 1);
      setResults([...acc]);
    }
    setRunning(false);
  };

  const doneAll = progressDone >= totalCount && totalCount > 0;
  const successCount = results.filter((r) => r.ok).length;
  const failCount = results.filter((r) => !r.ok).length;

  return (
    <Modal
      forceRender
      destroyOnClose
      visible={visible}
      title="批量设置发卡规则"
      width={640}
      maskClosable={!running}
      closable={!running}
      okButtonProps={{ disabled: !selectedRuleId || running, loading: running }}
      okText={doneAll ? '完成' : '开始执行'}
      cancelButtonProps={{ disabled: running }}
      onOk={() => {
        if (doneAll) {
          onSuccess();
        } else {
          void handleConfirm();
        }
      }}
      onCancel={onCancel}
    >
      <Alert
        showIcon
        type="info"
        style={{ marginBottom: 12 }}
        message="批量将选中商品改为「提货卡」履约模式并关联此规则"
        description="过程中会逐条更新商品，失败条目不影响其他条目。平台/租户级规则也能对商户商品生效。"
      />
      <Space direction="vertical" size={12} style={{ width: '100%' }}>
        <div>
          <Text strong>选中商品</Text>：共 {totalCount} 个
        </div>
        <div>
          <Text strong>选择规则</Text>
          <Select
            style={{ width: '100%', marginTop: 6 }}
            options={ruleOptions}
            value={selectedRuleId}
            onChange={setSelectedRuleId}
            placeholder={
              ruleOptions.length === 0
                ? '当前主体范围暂无启用规则，请先在「发卡规则管理」创建'
                : '请选择发卡规则'
            }
            disabled={running || ruleOptions.length === 0}
            showSearch
            optionFilterProp="label"
          />
        </div>
        {totalCount > 0 ? (
          <Progress
            percent={totalCount > 0 ? Math.round((progressDone / totalCount) * 100) : 0}
            status={failCount > 0 && doneAll ? 'exception' : doneAll ? 'success' : 'active'}
            format={() => `${progressDone} / ${totalCount}`}
          />
        ) : null}
        {doneAll ? (
          <Alert
            type={failCount === 0 ? 'success' : 'warning'}
            showIcon
            message={
              failCount === 0
                ? `全部成功：${successCount} 个`
                : `部分成功：成功 ${successCount} 个，失败 ${failCount} 个`
            }
            description={
              failCount > 0 ? (
                <div style={{ maxHeight: 120, overflowY: 'auto' }}>
                  {results
                    .filter((r) => !r.ok)
                    .map((r) => (
                      <div key={r.id}>
                        <Text type="danger">
                          {r.name}（#{r.id}）：{r.error}
                        </Text>
                      </div>
                    ))}
                </div>
              ) : null
            }
          />
        ) : null}
      </Space>
    </Modal>
  );
};

export default BatchSetFulfillmentRuleModal;
