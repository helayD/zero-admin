import React, { useEffect } from 'react';
import { Alert, Form, Modal } from 'antd';
import type { ProductSpuDraftFormValues } from '../data.d';
import { buildDraftClientValidationErrors, buildDraftFieldErrors } from '../draftFeedback';
import ProductSpuFormContent from './ProductSpuFormContent';
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

const formLayout = {
  labelCol: { span: 6 },
  wrapperCol: { span: 17 },
};

const UpdateModal: React.FC<UpdateModalProps> = (props) => {
  const [form] = Form.useForm();
  const { onSubmit, onCancel, updateVisible, currentData, scope, submitError } = props;

  useEffect(() => {
    if (form && !updateVisible) {
      form.resetFields();
    }
  }, [form, updateVisible]);

  useEffect(() => {
    if (currentData) {
      form.setFieldsValue({ ...currentData });
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

  return (
    <Modal
      forceRender
      destroyOnClose
      title="编辑商品"
      visible={updateVisible}
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
      <Form {...formLayout} form={form} onFinish={handleFinish}>
        <ProductSpuFormContent form={form} mode="update" visible={updateVisible} scope={scope} />
      </Form>
    </Modal>
  );
};

export default UpdateModal;
