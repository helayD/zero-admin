import React, { useEffect } from 'react';
import { Form, Input, InputNumber, Modal, Radio } from 'antd';
import type { PreferredAreaListItem } from '../data.d';

export interface UpdateFormProps {
  onCancel: () => void;
  onSubmit: (values: PreferredAreaListItem) => void;
  updateModalVisible: boolean;
  currentData: Partial<PreferredAreaListItem>;
}

const FormItem = Form.Item;

const formLayout = {
  labelCol: { span: 6 },
  wrapperCol: { span: 16 },
};

const UpdatePreferredAreaForm: React.FC<UpdateFormProps> = (props) => {
  const [form] = Form.useForm();
  const { onSubmit, onCancel, updateModalVisible, currentData } = props;

  useEffect(() => {
    if (form && !updateModalVisible) {
      form.resetFields();
    }
  }, [updateModalVisible, form]);

  useEffect(() => {
    if (currentData && updateModalVisible) {
      form.setFieldsValue({
        ...currentData,
      });
    }
  }, [currentData, updateModalVisible, form]);

  return (
    <Modal
      forceRender
      destroyOnClose
      title="编辑优选专区"
      visible={updateModalVisible}
      okText="保存"
      onOk={() => form.submit()}
      onCancel={onCancel}
      width={760}
    >
      <Form {...formLayout} form={form} onFinish={onSubmit}>
        <FormItem name="id" hidden>
          <Input />
        </FormItem>
        <FormItem name="name" label="专区名称" rules={[{ required: true, message: '请输入专区名称' }]}>
          <Input placeholder="请输入专区名称" />
        </FormItem>
        <FormItem name="subTitle" label="副标题">
          <Input placeholder="请输入副标题" />
        </FormItem>
        <FormItem name="pic" label="图片链接">
          <Input placeholder="请输入图片 URL" />
        </FormItem>
        <FormItem name="sort" label="排序" initialValue={0}>
          <InputNumber style={{ width: '100%' }} min={0} />
        </FormItem>
        <FormItem name="showStatus" label="显示状态">
          <Radio.Group>
            <Radio value={1}>显示</Radio>
            <Radio value={0}>隐藏</Radio>
          </Radio.Group>
        </FormItem>
      </Form>
    </Modal>
  );
};

export default UpdatePreferredAreaForm;
