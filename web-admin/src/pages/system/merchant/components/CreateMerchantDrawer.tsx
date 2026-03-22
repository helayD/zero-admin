import React from 'react';
import { Alert, message } from 'antd';
import {
  DrawerForm,
  ProFormCheckbox,
  ProFormDigit,
  ProFormGroup,
  ProFormText,
  ProFormTextArea,
} from '@ant-design/pro-form';
import { createMerchant } from '../service';
import type { CreateMerchantResult, MerchantCreateParams } from '../data.d';

type CreateMerchantDrawerProps = {
  visible: boolean;
  onVisibleChange: (visible: boolean) => void;
  onSuccess: (data: CreateMerchantResult) => void;
  channelOptions: { label: string; value: string }[];
  capabilityOptions: { label: string; value: string }[];
};

const readErrorMessage = (error: any, fallback: string) => {
  return error?.data?.message || error?.message || fallback;
};

const CreateMerchantDrawer: React.FC<CreateMerchantDrawerProps> = ({
  visible,
  onVisibleChange,
  onSuccess,
  channelOptions,
  capabilityOptions,
}) => {
  return (
    <DrawerForm<MerchantCreateParams>
      title="创建商户申请"
      width={680}
      visible={visible}
      drawerProps={{
        destroyOnClose: true,
        onClose: () => onVisibleChange(false),
      }}
      onVisibleChange={onVisibleChange}
      submitter={{
        searchConfig: {
          submitText: '提交申请',
          resetText: '重置',
        },
      }}
      initialValues={{
        availableChannels: ['app'],
        capabilityFlags: ['oms'],
        visibleScopeHint: '启用后会同步影响商户后台访问、能力包、菜单模板继承与作用域生效状态。',
      }}
      onFinish={async (values) => {
        try {
          const result = await createMerchant(values);
          message.success('商户申请已创建，当前处于待审核状态');
          onSuccess(result.data);
          return true;
        } catch (error) {
          message.error(readErrorMessage(error, '创建失败，请检查租户归属、商户编码或管理员信息'));
          return false;
        }
      }}
    >
      <Alert
        showIcon
        type="warning"
        message="Consequence Preview"
        description="提交后会写入商户主体真相源与治理审计。审核通过前不会直接开放商户后台访问，但会为后续菜单模板、能力包和商户管理员绑定预留稳定挂点。"
        style={{ marginBottom: 24 }}
      />

      <ProFormGroup>
        <ProFormDigit
          name="tenantId"
          label="归属租户 ID"
          width="md"
          fieldProps={{ precision: 0 }}
          rules={[{ required: true, message: '请输入归属租户 ID' }]}
        />
        <ProFormDigit
          name="primaryAdminUserId"
          label="商户管理员用户 ID"
          width="md"
          fieldProps={{ precision: 0 }}
          extra="可选。若已有后台管理员，可在这里提前绑定。"
        />
      </ProFormGroup>

      <ProFormGroup>
        <ProFormText
          name="merchantName"
          label="商户名称"
          width="md"
          rules={[{ required: true, message: '请输入商户名称' }]}
        />
        <ProFormText name="merchantShortName" label="商户简称" width="md" />
      </ProFormGroup>

      <ProFormGroup>
        <ProFormText name="merchantCode" label="商户编码" width="md" extra="留空时系统自动生成" />
        <ProFormText
          name="contactName"
          label="联系人"
          width="md"
          rules={[{ required: true, message: '请输入联系人' }]}
        />
      </ProFormGroup>

      <ProFormGroup>
        <ProFormText
          name="contactMobile"
          label="联系电话"
          width="md"
          rules={[{ required: true, message: '请输入联系电话' }]}
        />
        <ProFormText name="contactEmail" label="联系邮箱" width="md" />
      </ProFormGroup>

      <ProFormCheckbox.Group
        name="availableChannels"
        label="可用渠道"
        options={channelOptions}
        rules={[{ required: true, message: '请至少选择一个可用渠道' }]}
      />

      <ProFormCheckbox.Group
        name="capabilityFlags"
        label="能力包"
        options={capabilityOptions}
        rules={[{ required: true, message: '请至少选择一个能力包' }]}
        extra="这里保存的是主体治理侧的可用能力，不等于完整角色授权。"
      />

      <ProFormTextArea
        name="visibleScopeHint"
        label="可见范围提示"
        fieldProps={{ rows: 3 }}
        extra="用于后台详情和后续治理页提示当前商户作用域影响。"
      />

      <ProFormTextArea
        name="remark"
        label="备注"
        fieldProps={{ rows: 3 }}
        extra="可记录申请来源、线下核验备注或待跟进事项。"
      />
    </DrawerForm>
  );
};

export default CreateMerchantDrawer;
