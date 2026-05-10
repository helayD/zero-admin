import React, { useEffect, useMemo, useState } from 'react';
import {
  Alert,
  Descriptions,
  Form,
  Input,
  InputNumber,
  Modal,
  Radio,
  Select,
  Space,
  Typography,
  TreeSelect,
} from 'antd';
import { history } from 'umi';
import type { ProductSpuDraftFormValues } from '../data.d';
import { buildDraftClientValidationErrors, buildDraftFieldErrors } from '../draftFeedback';
import NestedDraftSections from './NestedDraftSections';
import { useCatalogOptions } from './useCatalogOptions';
import type { CatalogActionError } from '@/pages/pms/errorFeedback';
import type { GovernanceScopeValue } from '@/pages/system/components/governance';
import { toGovernancePayload } from '@/pages/system/components/governance';
import { queryProductFulfillmentRuleList } from '@/pages/sms/ProductFulfillmentRule/service';
import UploadFileComponents from '@/components/common/UploadFileComponents';

const { Link } = Typography;

// B3: 规则下拉 option 加厚，显示规则名+模板+有效期+可转赠
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

export interface AddModalProps {
  onCancel: () => void;
  onSubmit: (values: ProductSpuDraftFormValues) => void;
  addVisible: boolean;
  scope?: GovernanceScopeValue;
  submitError?: CatalogActionError;
}

const FormItem = Form.Item;

const formLayout = {
  labelCol: { span: 7 },
  wrapperCol: { span: 13 },
};

const initialValues: ProductSpuDraftFormValues = {
  publishStatus: 0,
  newStatus: 0,
  recommendStatus: 0,
  verifyStatus: 0,
  previewStatus: 0,
  promotionType: 0,
  fulfillmentMode: 'physical_delivery',
  skuList: [{ publishStatus: 0, verifyStatus: 0 }],
  memberPriceList: [],
  ladderList: [],
  fullList: [],
  attributeValueList: [],
};

const AddModal: React.FC<AddModalProps> = (props) => {
  const [form] = Form.useForm();
  const [fulfillmentMode, setFulfillmentMode] = useState<string>('physical_delivery');
  const [ruleInfos, setRuleInfos] = useState<RuleInfo[]>([]);
  const fulfillmentRuleId = Form.useWatch('fulfillmentRuleId', form);

  const { onSubmit, onCancel, addVisible, scope, submitError } = props;
  const {
    attributeOptions,
    brandOptions,
    categoryNameById,
    categoryPathById,
    categoryTreeOptions,
  } = useCatalogOptions(addVisible, scope);

  useEffect(() => {
    if (form && !addVisible) {
      form.resetFields();
    }
  }, [addVisible, form]);

  useEffect(() => {
    if (!addVisible || !scope) {
      setRuleInfos([]);
      return;
    }
    queryProductFulfillmentRuleList({
      pageSize: 999,
      ruleStatus: 1, // Story 10.10 Task 4.5: 只展示启用规则，避免运营选到已禁用规则
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
  }, [addVisible, scope]);

  const ruleOptions = useMemo(
    () => ruleInfos.map((rule) => ({ label: buildRuleOptionLabel(rule), value: rule.id })),
    [ruleInfos],
  );

  const selectedRule = useMemo(
    () => ruleInfos.find((rule) => rule.id === fulfillmentRuleId),
    [ruleInfos, fulfillmentRuleId],
  );

  useEffect(() => {
    if (!submitError) {
      return;
    }
    const fieldErrors = buildDraftFieldErrors(submitError.description, form.getFieldsValue(true));
    if (fieldErrors.length > 0) {
      form.setFields(fieldErrors);
    }
  }, [form, submitError]);

  const handleSubmit = () => {
    if (!form) return;
    form.submit();
  };

  const handleFinish = (values: ProductSpuDraftFormValues) => {
    const fieldErrors = buildDraftClientValidationErrors(values);
    if (fieldErrors.length > 0) {
      form.setFields(fieldErrors);
      return;
    }
    if (onSubmit) {
      onSubmit(values);
    }
  };

  const renderContent = () => {
    return (
      <>
        <FormItem name="id" hidden>
          <Input id="create-id" />
        </FormItem>
        <FormItem
          name="name"
          label="商品名称"
          rules={[{ required: true, message: '请输入商品名称!' }]}
        >
          <Input id="create-name" placeholder={'请输入商品名称!'} />
        </FormItem>
        <FormItem
          name="productSn"
          label="商品货号"
          rules={[{ required: true, message: '请输入商品货号!' }]}
        >
          <Input id="create-productSn" placeholder={'请输入商品货号!'} />
        </FormItem>
        <FormItem
          name="categoryId"
          label="商品分类"
          rules={[{ required: true, message: '请选择商品分类!' }]}
        >
          <TreeSelect
            showSearch
            treeDefaultExpandAll
            style={{ width: '100%' }}
            treeData={categoryTreeOptions}
            placeholder="请选择商品分类"
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
        <FormItem name="categoryIds" hidden>
          <Input id="create-categoryIds" />
        </FormItem>
        <FormItem name="categoryName" hidden>
          <Input id="create-categoryName" />
        </FormItem>
        <FormItem
          name="brandId"
          label="商品品牌"
          rules={[{ required: true, message: '请选择商品品牌!' }]}
        >
          <Select
            showSearch
            allowClear
            optionFilterProp="label"
            options={brandOptions}
            placeholder="请选择商品品牌"
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
        <FormItem name="brandName" hidden>
          <Input id="create-brandName" />
        </FormItem>
        {/* B1: 履约模式 + 发卡规则 前置到品牌后面的基础信息区 */}
        <FormItem
          name="fulfillmentMode"
          label="履约模式"
          rules={[{ required: true, message: '请选择履约模式!' }]}
          tooltip="实物发货：商品直接进入物流履约流程；提货卡：支付成功后生成提货卡资产并复用提货卡入账"
        >
          <Select
            options={fulfillmentModeOptions}
            placeholder="请选择履约模式"
            onChange={(value: string) => {
              setFulfillmentMode(value);
              if (value !== 'digital_asset') {
                form.setFieldsValue({ fulfillmentRuleId: undefined });
              }
            }}
          />
        </FormItem>
        {/* B2: 切换到提货卡模式 引导 Alert */}
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
            rules={[{ required: true, message: '请选择发卡规则!' }]}
            tooltip="提货卡模式下必须关联一个有效的发卡规则"
            extra={
              ruleOptions.length === 0 ? (
                // B4: 空态引导
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
              placeholder="请选择发卡规则"
              showSearch
              optionFilterProp="label"
              style={{ width: '100%' }}
              disabled={ruleOptions.length === 0}
            />
          </FormItem>
        )}
        {/* B5: 选中规则后展示只读关键属性卡片 */}
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
        <FormItem name="unit" label="单位" rules={[{ required: true, message: '请输入单位!' }]}>
          <Input id="create-unit" placeholder={'请输入单位!'} />
        </FormItem>
        <FormItem
          name="weight"
          label="重量(kg)"
          rules={[{ required: true, message: '请输入重量(kg)!' }]}
        >
          <Input id="create-weight" placeholder={'请输入重量(kg)!'} />
        </FormItem>
        <FormItem
          name="keywords"
          label="关键词"
          rules={[{ required: true, message: '请输入关键词!' }]}
        >
          <Input id="create-keywords" placeholder={'请输入关键词!'} />
        </FormItem>
        <FormItem
          name="albumPics"
          label="画册图片"
          rules={[{ required: true, message: '请上传画册图片!' }]}
        >
          <UploadFileComponents count={8} />
        </FormItem>
        <FormItem name="mainPic" label="主图" rules={[{ required: true, message: '请上传主图!' }]}>
          <UploadFileComponents count={1} />
        </FormItem>
        <FormItem
          name="publishStatus"
          label="上架状态：0-下架，1-上架"
          rules={[{ required: true, message: '请输入上架状态：0-下架，1-上架!' }]}
        >
          <Radio.Group>
            <Radio value={0}>下架</Radio>
            <Radio value={1}>上架</Radio>
          </Radio.Group>
        </FormItem>
        <FormItem
          name="newStatus"
          label="新品状态:0->不是新品；1->新品"
          rules={[{ required: true, message: '请输入新品状态:0->不是新品；1->新品!' }]}
        >
          <Radio.Group>
            <Radio value={0}>否</Radio>
            <Radio value={1}>是</Radio>
          </Radio.Group>
        </FormItem>
        <FormItem
          name="recommendStatus"
          label="推荐状态；0->不推荐；1->推荐"
          rules={[{ required: true, message: '请输入推荐状态；0->不推荐；1->推荐!' }]}
        >
          <Radio.Group>
            <Radio value={0}>否</Radio>
            <Radio value={1}>是</Radio>
          </Radio.Group>
        </FormItem>
        <FormItem
          name="verifyStatus"
          label="审核状态：0->未审核；1->审核通过"
          rules={[{ required: true, message: '请输入审核状态：0->未审核；1->审核通过!' }]}
        >
          <Radio.Group>
            <Radio value={0}>未审核</Radio>
            <Radio value={1}>审核通过</Radio>
          </Radio.Group>
        </FormItem>
        <FormItem
          name="previewStatus"
          label="是否为预告商品：0->不是；1->是"
          rules={[{ required: true, message: '请输入是否为预告商品：0->不是；1->是!' }]}
        >
          <Radio.Group>
            <Radio value={0}>否</Radio>
            <Radio value={1}>是</Radio>
          </Radio.Group>
        </FormItem>
        <FormItem name="sort" label="排序" rules={[{ required: true, message: '请输入排序!' }]}>
          <InputNumber style={{ width: 255 }} />
        </FormItem>
        <FormItem
          name="newStatusSort"
          label="新品排序"
          rules={[{ required: true, message: '请输入新品排序!' }]}
        >
          <InputNumber style={{ width: 255 }} />
        </FormItem>
        <FormItem
          name="recommendStatusSort"
          label="推荐排序"
          rules={[{ required: true, message: '请输入推荐排序!' }]}
        >
          <InputNumber style={{ width: 255 }} />
        </FormItem>
        <FormItem name="stock" label="库存" rules={[{ required: true, message: '请输入库存!' }]}>
          <Input id="create-stock" placeholder={'请输入库存!'} />
        </FormItem>
        <FormItem
          name="lowStock"
          label="预警库存"
          rules={[{ required: true, message: '请输入预警库存!' }]}
        >
          <Input id="create-lowStock" placeholder={'请输入预警库存!'} />
        </FormItem>
        <FormItem
          name="promotionType"
          label="促销类型：0->没有促销使用原价;1->使用促销价；2->使用会员价；3->使用阶梯价格；4->使用满减价格；5->秒杀"
          rules={[
            {
              required: true,
              message:
                '请输入促销类型：0->没有促销使用原价;1->使用促销价；2->使用会员价；3->使用阶梯价格；4->使用满减价格；5->秒杀!',
            },
          ]}
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
        <FormItem name="subTitle" label="副标题">
          <Input id="create-subTitle" placeholder={'请输入副标题!'} />
        </FormItem>
        <FormItem
          name="detailHtml"
          label="产品详情网页内容"
          rules={[{ required: true, message: '请输入产品详情网页内容!' }]}
        >
          <Input id="create-detailHtml" placeholder={'请输入产品详情网页内容!'} />
        </FormItem>
        <FormItem
          name="detailMobileHtml"
          label="移动端网页详情"
          rules={[{ required: true, message: '请输入移动端网页详情!' }]}
        >
          <Input id="create-detailMobileHtml" placeholder={'请输入移动端网页详情!'} />
        </FormItem>
        <NestedDraftSections attributeOptions={attributeOptions} />
      </>
    );
  };

  const modalFooter = { okText: '保存', onOk: handleSubmit, onCancel };

  return (
    <Modal
      forceRender
      destroyOnClose
      title="新增"
      visible={addVisible}
      width={960}
      bodyStyle={{ maxHeight: '72vh', overflowY: 'auto' }}
      {...modalFooter}
    >
      {submitError && (
        <Alert
          showIcon
          type="error"
          style={{ marginBottom: 16 }}
          message={submitError.title}
          description={submitError.description}
        />
      )}
      <Form {...formLayout} form={form} initialValues={initialValues} onFinish={handleFinish}>
        {renderContent()}
      </Form>
    </Modal>
  );
};

export default AddModal;
