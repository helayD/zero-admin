import React from 'react';
import { Alert, message } from 'antd';
import {
  DrawerForm,
  ProFormCheckbox,
  ProFormDigit,
  ProFormGroup,
  ProFormText,
} from '@ant-design/pro-form';
import { createTenant } from '../service';
import type { CreateTenantResult, TenantCreateParams } from '../data.d';

type CreateTenantDrawerProps = {
  visible: boolean;
  onVisibleChange: (visible: boolean) => void;
  onSuccess: (data: CreateTenantResult) => void;
  channelOptions: { label: string; value: string }[];
  featureOptions: { label: string; value: string }[];
};

const readErrorMessage = (error: any, fallback: string) => {
  return error?.data?.message || error?.message || fallback;
};

const CreateTenantDrawer: React.FC<CreateTenantDrawerProps> = ({
  visible,
  onVisibleChange,
  onSuccess,
  channelOptions,
  featureOptions,
}) => {
  return (
    <DrawerForm<TenantCreateParams>
      title="创建租户"
      width={640}
      visible={visible}
      drawerProps={{
        destroyOnClose: true,
        onClose: () => onVisibleChange(false),
      }}
      onVisibleChange={onVisibleChange}
      submitter={{
        searchConfig: {
          submitText: '确认创建',
          resetText: '重置',
        },
      }}
      initialValues={{
        availableChannels: ['app'],
        dataRetentionDays: 180,
        featureFlags: [],
      }}
      onFinish={async (values) => {
        try {
          const result = await createTenant(values);
          message.success('租户创建成功，已进入待激活状态');
          onSuccess(result.data);
          return true;
        } catch (error) {
          message.error(
            readErrorMessage(error, '创建失败，请检查管理员账号是否已被绑定或租户名称是否重复'),
          );
          return false;
        }
      }}
    >
      <Alert
        showIcon
        type="warning"
        message="影响预览"
        description="提交后会同时创建租户主体、首个管理员账号、bootstrap 作用域元数据，并写入开通审计记录。管理员默认保持待激活，不会直接获得平台全量权限。"
        style={{ marginBottom: 24 }}
      />

      <ProFormGroup>
        <ProFormText
          name="tenantName"
          label="租户名称"
          width="md"
          rules={[{ required: true, message: '请输入租户名称' }]}
        />
        <ProFormText name="tenantShortName" label="租户简称" width="md" />
      </ProFormGroup>

      <ProFormGroup>
        <ProFormText
          name="contactName"
          label="联系人"
          width="md"
          rules={[{ required: true, message: '请输入联系人' }]}
        />
        <ProFormText
          name="contactMobile"
          label="联系电话"
          width="md"
          rules={[{ required: true, message: '请输入联系电话' }]}
        />
      </ProFormGroup>

      <ProFormText name="contactEmail" label="联系邮箱" width="xl" />

      <ProFormCheckbox.Group
        name="availableChannels"
        label="可用渠道"
        options={channelOptions}
        rules={[{ required: true, message: '请至少选择一个渠道' }]}
      />

      <ProFormDigit
        name="dataRetentionDays"
        label="数据保留天数"
        min={30}
        max={3650}
        fieldProps={{ precision: 0 }}
        rules={[{ required: true, message: '请输入数据保留天数' }]}
      />

      <ProFormCheckbox.Group
        name="featureFlags"
        label="业务开关"
        options={featureOptions}
        extra="可按当前开通范围先做最小治理，后续再在租户侧继续扩展开关。"
      />

      <Alert
        showIcon
        type="info"
        message="首个管理员初始化"
        description="这里创建的是受控 bootstrap 账号。若提交失败，请优先检查管理员手机号、账号是否已被其他租户绑定。"
        style={{ marginBottom: 24 }}
      />

      <ProFormGroup>
        <ProFormText
          name="adminUserName"
          label="管理员账号"
          width="md"
          rules={[{ required: true, message: '请输入管理员账号' }]}
        />
        <ProFormText
          name="adminNickName"
          label="管理员昵称"
          width="md"
          rules={[{ required: true, message: '请输入管理员昵称' }]}
        />
      </ProFormGroup>

      <ProFormGroup>
        <ProFormText
          name="adminMobile"
          label="管理员手机号"
          width="md"
          rules={[{ required: true, message: '请输入管理员手机号' }]}
        />
        <ProFormText name="adminEmail" label="管理员邮箱" width="md" />
      </ProFormGroup>

      <ProFormText.Password
        name="adminPassword"
        label="初始化密码"
        width="md"
        rules={[{ required: true, message: '请输入初始化密码' }]}
      />
    </DrawerForm>
  );
};

export default CreateTenantDrawer;
