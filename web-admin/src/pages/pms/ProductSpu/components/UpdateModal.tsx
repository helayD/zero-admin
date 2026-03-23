import React, { useEffect } from 'react';
import { Alert, Form, Input, InputNumber, Modal, Radio, Select, TreeSelect } from 'antd';
import type { ProductSpuDraftFormValues } from '../data.d';
import { buildDraftClientValidationErrors, buildDraftFieldErrors } from '../draftFeedback';
import NestedDraftSections from './NestedDraftSections';
import { useCatalogOptions } from './useCatalogOptions';
import type { CatalogActionError } from '@/pages/pms/errorFeedback';
import type { GovernanceScopeValue } from '@/pages/system/components/governance';

export interface UpdateModalProps {
  onCancel: () => void;
  onSubmit: (values: ProductSpuDraftFormValues) => void;
  updateVisible: boolean;
  currentData: ProductSpuDraftFormValues;
  scope?: GovernanceScopeValue;
  submitError?: CatalogActionError;
}

const FormItem = Form.Item;

const formLayout = {
  labelCol: { span: 7 },
  wrapperCol: { span: 13 },
};

const UpdateModal: React.FC<UpdateModalProps> = (props) => {
  const [form] = Form.useForm();

  const { onSubmit, onCancel, updateVisible, currentData, scope, submitError } = props;
  const {
    attributeOptions,
    brandOptions,
    categoryNameById,
    categoryPathById,
    categoryTreeOptions,
  } = useCatalogOptions(updateVisible, scope);

  useEffect(() => {
    if (form && !updateVisible) {
      form.resetFields();
    }
  }, [form, updateVisible]);

  useEffect(() => {
    if (currentData) {
      form.setFieldsValue({
        ...currentData,
      });
    }
  }, [currentData, form]);

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
        <FormItem name="id" label="主键" hidden>
          <Input id="update-id" />
        </FormItem>
        <FormItem
          name="name"
          label="商品名称"
          rules={[{ required: true, message: '请输入商品名称!' }]}
        >
          <Input id="update-name" placeholder={'请输入商品名称!'} />
        </FormItem>
        <FormItem
          name="productSn"
          label="商品货号"
          rules={[{ required: true, message: '请输入商品货号!' }]}
        >
          <Input id="update-productSn" placeholder={'请输入商品货号!'} />
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
          <Input id="update-categoryIds" />
        </FormItem>
        <FormItem name="categoryName" hidden>
          <Input id="update-categoryName" />
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
          <Input id="update-brandName" />
        </FormItem>
        <FormItem name="unit" label="单位" rules={[{ required: true, message: '请输入单位!' }]}>
          <Input id="update-unit" placeholder={'请输入单位!'} />
        </FormItem>
        <FormItem
          name="weight"
          label="重量(kg)"
          rules={[{ required: true, message: '请输入重量(kg)!' }]}
        >
          <Input id="update-weight" placeholder={'请输入重量(kg)!'} />
        </FormItem>
        <FormItem
          name="keywords"
          label="关键词"
          rules={[{ required: true, message: '请输入关键词!' }]}
        >
          <Input id="update-keywords" placeholder={'请输入关键词!'} />
        </FormItem>
        <FormItem
          name="albumPics"
          label="画册图片，最多8张，以逗号分割"
          rules={[{ required: true, message: '请输入画册图片，最多8张，以逗号分割!' }]}
        >
          <Input id="update-albumPics" placeholder={'请输入画册图片，最多8张，以逗号分割!'} />
        </FormItem>
        <FormItem name="mainPic" label="主图" rules={[{ required: true, message: '请输入主图!' }]}>
          <Input id="update-mainPic" placeholder={'请输入主图!'} />
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
          <Input id="update-stock" placeholder={'请输入库存!'} />
        </FormItem>
        <FormItem
          name="lowStock"
          label="预警库存"
          rules={[{ required: true, message: '请输入预警库存!' }]}
        >
          <Input id="update-lowStock" placeholder={'请输入预警库存!'} />
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
          <Input id="update-subTitle" placeholder={'请输入副标题!'} />
        </FormItem>
        <FormItem
          name="detailHtml"
          label="产品详情网页内容"
          rules={[{ required: true, message: '请输入产品详情网页内容!' }]}
        >
          <Input id="update-detailHtml" placeholder={'请输入产品详情网页内容!'} />
        </FormItem>
        <FormItem
          name="detailMobileHtml"
          label="移动端网页详情"
          rules={[{ required: true, message: '请输入移动端网页详情!' }]}
        >
          <Input id="update-detailMobileHtml" placeholder={'请输入移动端网页详情!'} />
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
      title="编辑"
      visible={updateVisible}
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
      <Form {...formLayout} form={form} onFinish={handleFinish}>
        {renderContent()}
      </Form>
    </Modal>
  );
};

export default UpdateModal;
