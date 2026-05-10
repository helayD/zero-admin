import React, { useEffect, useMemo, useState } from 'react';
import {
  Alert,
  Collapse,
  Descriptions,
  Form,
  Input,
  InputNumber,
  Radio,
  Select,
  Space,
  TreeSelect,
  Typography,
} from 'antd';
import type { FormInstance } from 'antd';
import { history } from 'umi';
import NestedDraftSections from './NestedDraftSections';
import { useCatalogOptions } from './useCatalogOptions';
import type { GovernanceScopeValue } from '@/pages/system/components/governance';
import { toGovernancePayload } from '@/pages/system/components/governance';
import { queryProductFulfillmentRuleList } from '@/pages/sms/ProductFulfillmentRule/service';
import UploadFileComponents from '@/components/common/UploadFileComponents';

const { Link } = Typography;
const { Panel } = Collapse;

type RuleInfo = {
  id: number;
  ruleName: string;
  cardTemplateName?: string;
  expireDays?: number;
  transferable?: number;
  transferLimit?: number;
};

const buildRuleOptionLabel = (rule: RuleInfo) => {
  const parts: string[] = [rule.ruleName];
  if (rule.cardTemplateName) parts.push(rule.cardTemplateName);
  if (typeof rule.expireDays === 'number' && rule.expireDays > 0) {
    parts.push(`${rule.expireDays}天`);
  }
  if (rule.transferable === 1) {
    parts.push(
      rule.transferLimit && rule.transferLimit > 0
        ? `可转赠·限${rule.transferLimit}次`
        : '可转赠',
    );
  } else {
    parts.push('不可转赠');
  }
  return parts.join(' · ');
};

const fulfillmentModeOptions = [
  { value: 'physical_delivery', label: '实物发货' },
  { value: 'digital_asset', label: '提货卡（支付后生成提货卡入账）' },
];

export type ProductSpuFormMode = 'add' | 'update';

export interface ProductSpuFormContentProps {
  form: FormInstance;
  mode: ProductSpuFormMode;
  visible: boolean;
  scope?: GovernanceScopeValue;
}

const FormItem = Form.Item;

const ProductSpuFormContent: React.FC<ProductSpuFormContentProps> = ({
  form,
  mode,
  visible,
  scope,
}) => {
  const [ruleInfos, setRuleInfos] = useState<RuleInfo[]>([]);
  const fulfillmentMode = Form.useWatch('fulfillmentMode', form) ?? 'physical_delivery';
  const fulfillmentRuleId = Form.useWatch('fulfillmentRuleId', form);

  const {
    attributeOptions,
    brandOptions,
    categoryNameById,
    categoryPathById,
    categoryTreeOptions,
  } = useCatalogOptions(visible, scope);

  useEffect(() => {
    if (!visible || !scope) {
      setRuleInfos([]);
      return;
    }
    queryProductFulfillmentRuleList({
      pageSize: 999,
      ruleStatus: 1,
      ...toGovernancePayload(scope),
    })
      .then((res) => {
        const infos: RuleInfo[] = (res.data || []).map((item: any) => ({
          id: item.id,
          ruleName: item.ruleName,
          cardTemplateName: item.cardTemplateName,
          expireDays: item.expireDays,
          transferable: item.transferable,
          transferLimit: item.transferLimit,
        }));
        setRuleInfos(infos);
      })
      .catch(() => {
        setRuleInfos([]);
      });
  }, [visible, scope]);

  const ruleOptions = useMemo(
    () => ruleInfos.map((rule) => ({ label: buildRuleOptionLabel(rule), value: rule.id })),
    [ruleInfos],
  );

  const selectedRule = useMemo(
    () => ruleInfos.find((rule) => rule.id === fulfillmentRuleId),
    [ruleInfos, fulfillmentRuleId],
  );

  const idPrefix = mode === 'add' ? 'create' : 'update';

  return (
    <>
      {/* 隐藏字段：默认值由外层 initialValues 提供，确保 service 拿到完整 payload */}
      <FormItem name="id" hidden>
        <Input id={`${idPrefix}-id`} />
      </FormItem>
      <FormItem name="categoryIds" hidden>
        <Input />
      </FormItem>
      <FormItem name="categoryName" hidden>
        <Input />
      </FormItem>
      <FormItem name="brandName" hidden>
        <Input />
      </FormItem>
      {/* 状态类字段由列表页的「更多」菜单管理，表单里只保留默认值 */}
      <FormItem name="publishStatus" hidden>
        <Input />
      </FormItem>
      <FormItem name="newStatus" hidden>
        <Input />
      </FormItem>
      <FormItem name="recommendStatus" hidden>
        <Input />
      </FormItem>
      <FormItem name="verifyStatus" hidden>
        <Input />
      </FormItem>
      <FormItem name="previewStatus" hidden>
        <Input />
      </FormItem>
      <FormItem name="newStatusSort" hidden>
        <Input />
      </FormItem>
      <FormItem name="recommendStatusSort" hidden>
        <Input />
      </FormItem>

      <Collapse
        defaultActiveKey={['basic', 'fulfillment', 'media']}
        bordered={false}
        style={{ background: 'transparent' }}
      >
        {/* ========== 基础信息 ========== */}
        <Panel header="基础信息" key="basic" forceRender>
          <FormItem
            name="name"
            label="商品名称"
            rules={[{ required: true, message: '请输入商品名称' }]}
          >
            <Input
              id={`${idPrefix}-name`}
              placeholder="如 iPhone 16 Pro Max 1TB 钛金蓝"
              maxLength={120}
              showCount
            />
          </FormItem>
          <FormItem
            name="productSn"
            label="商品货号"
            rules={[{ required: true, message: '请输入商品货号' }]}
            tooltip="内部 SKU/SPU 编码，用于库存对账、ERP 对接"
          >
            <Input
              id={`${idPrefix}-productSn`}
              placeholder="如 APL-IP16PM-1TB-BLU"
              maxLength={64}
            />
          </FormItem>
          <FormItem
            name="categoryId"
            label="商品分类"
            rules={[{ required: true, message: '请选择商品分类' }]}
          >
            <TreeSelect
              showSearch
              treeDefaultExpandAll
              style={{ width: '100%' }}
              treeData={categoryTreeOptions}
              placeholder="搜索或选择商品分类"
              onChange={(value) => {
                const categoryId = Number(value || 0);
                if (!categoryId) {
                  form.setFieldsValue({
                    categoryId: undefined,
                    categoryIds: undefined,
                    categoryName: undefined,
                  });
                  return;
                }
                form.setFieldsValue({
                  categoryId,
                  categoryIds: (categoryPathById[categoryId] || [categoryId]).join(','),
                  categoryName: categoryNameById[categoryId],
                });
              }}
            />
          </FormItem>
          <FormItem
            name="brandId"
            label="商品品牌"
            rules={[{ required: true, message: '请选择商品品牌' }]}
          >
            <Select
              showSearch
              allowClear
              optionFilterProp="label"
              options={brandOptions}
              placeholder="搜索或选择商品品牌"
              onChange={(value) => {
                const brandId = Number(value || 0);
                const brandOption = brandOptions.find((item) => item.value === brandId);
                form.setFieldsValue({
                  brandId: brandId || undefined,
                  brandName: brandOption?.label,
                });
              }}
            />
          </FormItem>
          <FormItem
            name="unit"
            label="单位"
            rules={[{ required: true, message: '请输入单位' }]}
          >
            <Input id={`${idPrefix}-unit`} placeholder="如 件 / 盒 / 台 / 张" maxLength={16} />
          </FormItem>
          <FormItem name="weight" label="重量(kg)" tooltip="实物发货建议填写，数字商品可留空">
            <InputNumber
              min={0}
              step={0.01}
              precision={2}
              style={{ width: '100%' }}
              placeholder="可选，单位 kg"
            />
          </FormItem>
          <FormItem name="subTitle" label="副标题" tooltip="展示在商品卡片副标题位">
            <Input
              id={`${idPrefix}-subTitle`}
              placeholder="可选，50 字以内"
              maxLength={80}
              showCount
            />
          </FormItem>
          <FormItem name="keywords" label="关键词" tooltip="SEO 用，多个词用英文逗号分隔">
            <Input
              id={`${idPrefix}-keywords`}
              placeholder="可选，如 iPhone,苹果手机,旗舰机"
              maxLength={200}
            />
          </FormItem>
          <FormItem name="sort" label="排序" tooltip="数值越大越靠前，默认 0">
            <InputNumber min={0} style={{ width: 200 }} placeholder="默认 0" />
          </FormItem>
        </Panel>

        {/* ========== 履约配置 ========== */}
        <Panel header="履约配置" key="fulfillment" forceRender>
          <FormItem
            name="fulfillmentMode"
            label="履约模式"
            rules={[{ required: true, message: '请选择履约模式' }]}
            tooltip="实物发货：下单后走物流履约；提货卡：下单后生成提货卡资产，用户发起提货时再走履约"
          >
            <Select
              options={fulfillmentModeOptions}
              placeholder="请选择履约模式"
              onChange={(value: string) => {
                if (value !== 'digital_asset') {
                  form.setFieldsValue({ fulfillmentRuleId: undefined });
                }
              }}
            />
          </FormItem>
          {fulfillmentMode === 'digital_asset' && (
            <FormItem label=" " colon={false}>
              <Alert
                type="info"
                showIcon
                message="该商品将以「提货卡」方式发放"
                description={
                  <Space direction="vertical" size={2}>
                    <span>用户下单后不进入物流流程，而是在「数字卡包」中生成卡片，待用户发起提货后再走履约。</span>
                    <span>请确认已在「发卡规则管理」配置好适用规则。</span>
                  </Space>
                }
              />
            </FormItem>
          )}
          {fulfillmentMode === 'digital_asset' && (
            <FormItem
              name="fulfillmentRuleId"
              label="发卡规则"
              rules={[{ required: true, message: '请选择发卡规则' }]}
              tooltip="提货卡模式下必须关联一个有效的发卡规则"
              extra={
                ruleOptions.length === 0 ? (
                  <Space size={4}>
                    <span style={{ color: '#ff4d4f' }}>当前主体范围暂无启用发卡规则，</span>
                    <Link onClick={() => history.push('/sms/ProductFulfillmentRule/list')}>
                      去创建
                    </Link>
                  </Space>
                ) : null
              }
            >
              <Select
                options={ruleOptions}
                placeholder="搜索或选择发卡规则"
                showSearch
                optionFilterProp="label"
                style={{ width: '100%' }}
                disabled={ruleOptions.length === 0}
              />
            </FormItem>
          )}
          {fulfillmentMode === 'digital_asset' && selectedRule && (
            <FormItem label=" " colon={false}>
              <Descriptions
                size="small"
                bordered
                column={2}
                labelStyle={{ background: '#fafafa', width: 110 }}
              >
                <Descriptions.Item label="规则名称">{selectedRule.ruleName}</Descriptions.Item>
                <Descriptions.Item label="卡片模板">
                  {selectedRule.cardTemplateName || '-'}
                </Descriptions.Item>
                <Descriptions.Item label="有效期">
                  {typeof selectedRule.expireDays === 'number' && selectedRule.expireDays > 0
                    ? `${selectedRule.expireDays} 天`
                    : '永久'}
                </Descriptions.Item>
                <Descriptions.Item label="是否可转赠">
                  {selectedRule.transferable === 1
                    ? selectedRule.transferLimit && selectedRule.transferLimit > 0
                      ? `可转赠（限 ${selectedRule.transferLimit} 次）`
                      : '可转赠'
                    : '不可转赠'}
                </Descriptions.Item>
              </Descriptions>
            </FormItem>
          )}
        </Panel>

        {/* ========== 媒体资源 ========== */}
        <Panel header="媒体资源" key="media" forceRender>
          <FormItem
            name="mainPic"
            label="主图"
            rules={[{ required: true, message: '请上传主图' }]}
            tooltip="建议 800×800，JPG/PNG 小于 2MB"
          >
            <UploadFileComponents count={1} />
          </FormItem>
          <FormItem name="albumPics" label="画册图片" tooltip="可选，最多 8 张，首张会作为商品轮播首图">
            <UploadFileComponents count={8} />
          </FormItem>
          <FormItem name="detailHtml" label="PC 端详情">
            <Input.TextArea
              id={`${idPrefix}-detailHtml`}
              rows={4}
              placeholder="可选，粘贴 HTML 片段；后续可在「详情富文本编辑」中维护"
            />
          </FormItem>
          <FormItem name="detailMobileHtml" label="移动端详情">
            <Input.TextArea
              id={`${idPrefix}-detailMobileHtml`}
              rows={4}
              placeholder="可选，移动端展示的 HTML 片段"
            />
          </FormItem>
        </Panel>

        {/* ========== 销售设置 ========== */}
        <Panel header="销售设置" key="sales" forceRender>
          <FormItem
            name="promotionType"
            label="促销类型"
            rules={[{ required: true, message: '请选择促销类型' }]}
            tooltip="决定前台展示的价格计算方式，默认原价"
          >
            <Radio.Group>
              <Radio value={0}>原价</Radio>
              <Radio value={1}>促销价</Radio>
              <Radio value={2}>会员价</Radio>
              <Radio value={3}>阶梯价</Radio>
              <Radio value={4}>满减价</Radio>
              <Radio value={5}>秒杀价</Radio>
            </Radio.Group>
          </FormItem>
          <FormItem
            name="stock"
            label="库存"
            tooltip="SPU 整体库存，实际下单会取 SKU 层库存，这里为兜底值"
          >
            <InputNumber min={0} style={{ width: 200 }} placeholder="默认 0" />
          </FormItem>
          <FormItem name="lowStock" label="预警库存" tooltip="低于此值会触发库存预警">
            <InputNumber min={0} style={{ width: 200 }} placeholder="默认 0" />
          </FormItem>
        </Panel>

        {/* ========== 规格 & 属性 ========== */}
        <Panel header="规格 & 属性 / SKU 建档" key="spec" forceRender>
          <NestedDraftSections attributeOptions={attributeOptions} />
        </Panel>
      </Collapse>
    </>
  );
};

export default ProductSpuFormContent;
