import React, { useEffect } from 'react';
import { Alert, Form, Modal } from 'antd';
import type { ProductSpuDraftFormValues } from '../data.d';
import { buildDraftClientValidationErrors, buildDraftFieldErrors } from '../draftFeedback';
import ProductSpuFormContent from './ProductSpuFormContent';
import type { CatalogActionError } from '@/pages/pms/errorFeedback';
import type { GovernanceScopeValue } from '@/pages/system/components/governance';

export interface AddModalProps {
  onCancel: () => void;
  onSubmit: (values: ProductSpuDraftFormValues) => void;
  addVisible: boolean;
  scope?: GovernanceScopeValue;
  submitError?: CatalogActionError;
}

const formLayout = {
  labelCol: { span: 6 },
  wrapperCol: { span: 17 },
};

// 新建时的默认值：状态字段全部默认为草稿/未审核/下架/不推荐/非新品/非预告，
// 这些字段已经从 UI 移除，状态变更统一走列表页「更多」菜单
const initialValues: ProductSpuDraftFormValues = {
  publishStatus: 0,
  newStatus: 0,
  recommendStatus: 0,
  verifyStatus: 0,
  previewStatus: 0,
  promotionType: 0,
  newStatusSort: 0,
  recommendStatusSort: 0,
  sort: 0,
  stock: 0,
  lowStock: 0,
  weight: 0,
  fulfillmentMode: 'physical_delivery',
  skuList: [{ publishStatus: 0, verifyStatus: 0 }],
  memberPriceList: [],
  ladderList: [],
  fullList: [],
  attributeValueList: [],
};

const AddModal: React.FC<AddModalProps> = (props) => {
  const [form] = Form.useForm();
  const { onSubmit, onCancel, addVisible, scope, submitError } = props;

  useEffect(() => {
    if (form && !addVisible) {
      form.resetFields();
    }
  }, [addVisible, form]);

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

  return (
    <Modal
      forceRender
      destroyOnClose
      title="新建商品"
      visible={addVisible}
      width={960}
      bodyStyle={{ maxHeight: '72vh', overflowY: 'auto' }}
      okText="保存"
      onOk={handleSubmit}
      onCancel={onCancel}
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
        <ProductSpuFormContent form={form} mode="add" visible={addVisible} scope={scope} />
      </Form>
    </Modal>
  );
};

export default AddModal;
