import {
  CloudUploadOutlined,
  MessageOutlined,
  NotificationOutlined,
  SaveOutlined,
} from '@ant-design/icons';
import { Button, Card, Form, Input, InputNumber, message, Space, Spin, Switch, Tabs } from 'antd';
import React, { useCallback, useEffect, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import type { SystemConfigData } from './data.d';
import { querySystemConfig, saveSystemConfig } from './service';
import styles from './index.less';

const { TabPane } = Tabs;

const readErrorMessage = (error: any, fallback: string) => {
  return error?.data?.message || error?.message || fallback;
};

const secretPlaceholder = (configured?: boolean) => {
  return configured ? '已配置，留空则不修改' : '请输入密钥';
};

const SystemConfigPage: React.FC = () => {
  const [form] = Form.useForm<SystemConfigData>();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [secretState, setSecretState] = useState({
    ossSecret: false,
    smsSecret: false,
    pushSecret: false,
  });

  const loadConfig = useCallback(async () => {
    setLoading(true);
    try {
      const result = await querySystemConfig();
      form.setFieldsValue(result.data);
      setSecretState({
        ossSecret: !!result.data?.oss?.accessKeySecretConfigured,
        smsSecret: !!result.data?.sms?.accessKeySecretConfigured,
        pushSecret: !!result.data?.push?.appSecretConfigured,
      });
    } catch (error) {
      message.error(readErrorMessage(error, '加载系统配置失败'));
    } finally {
      setLoading(false);
    }
  }, [form]);

  useEffect(() => {
    void loadConfig();
  }, [loadConfig]);

  const handleSave = async () => {
    try {
      const values = await form.validateFields();
      setSaving(true);
      await saveSystemConfig(values);
      message.success('保存成功');
      await loadConfig();
    } catch (error) {
      if (error?.errorFields) {
        return;
      }
      message.error(readErrorMessage(error, '保存系统配置失败'));
    } finally {
      setSaving(false);
    }
  };

  return (
    <PageContainer>
      <Card bordered={false} className={styles.configCard}>
        <Spin spinning={loading}>
          <Form form={form} layout="vertical">
            <Space direction="vertical" size={16} className={styles.formStack}>
              <Tabs defaultActiveKey="oss">
                <TabPane
                  key="oss"
                  tab={
                    <span>
                      <CloudUploadOutlined />
                      OSS
                    </span>
                  }
                >
                  <div className={styles.formGrid}>
                    <Form.Item
                      label="Endpoint"
                      name={['oss', 'endpoint']}
                      rules={[{ required: true, message: '请输入 OSS Endpoint' }]}
                    >
                      <Input placeholder="oss-cn-shenzhen.aliyuncs.com" />
                    </Form.Item>
                    <Form.Item
                      label="Bucket"
                      name={['oss', 'bucketName']}
                      rules={[{ required: true, message: '请输入 Bucket' }]}
                    >
                      <Input placeholder="mbjq" />
                    </Form.Item>
                    <Form.Item
                      label="AccessKey ID"
                      name={['oss', 'accessKeyId']}
                      rules={[{ required: true, message: '请输入 AccessKey ID' }]}
                    >
                      <Input autoComplete="off" />
                    </Form.Item>
                    <Form.Item label="AccessKey Secret" name={['oss', 'accessKeySecret']}>
                      <Input.Password
                        autoComplete="new-password"
                        placeholder={secretPlaceholder(secretState.ossSecret)}
                      />
                    </Form.Item>
                    <Form.Item
                      label="公开访问域名"
                      name={['oss', 'url']}
                      rules={[{ required: true, message: '请输入公开访问域名' }]}
                    >
                      <Input placeholder="https://speed.maibanjk.com/" />
                    </Form.Item>
                    <Form.Item
                      label="上传上限 MB"
                      name={['oss', 'maxSizeMb']}
                      rules={[{ required: true, message: '请输入上传大小上限' }]}
                    >
                      <InputNumber min={1} max={2048} className={styles.fullWidth} />
                    </Form.Item>
                  </div>
                </TabPane>

                <TabPane
                  key="sms"
                  tab={
                    <span>
                      <MessageOutlined />
                      短信
                    </span>
                  }
                >
                  <div className={styles.formGrid}>
                    <Form.Item label="启用" name={['sms', 'enabled']} valuePropName="checked">
                      <Switch />
                    </Form.Item>
                    <Form.Item label="服务商" name={['sms', 'provider']}>
                      <Input placeholder="aliyun" />
                    </Form.Item>
                    <Form.Item label="Endpoint" name={['sms', 'endpoint']}>
                      <Input placeholder="dysmsapi.aliyuncs.com" />
                    </Form.Item>
                    <Form.Item label="AccessKey ID" name={['sms', 'accessKeyId']}>
                      <Input autoComplete="off" />
                    </Form.Item>
                    <Form.Item label="AccessKey Secret" name={['sms', 'accessKeySecret']}>
                      <Input.Password
                        autoComplete="new-password"
                        placeholder={secretPlaceholder(secretState.smsSecret)}
                      />
                    </Form.Item>
                    <Form.Item label="短信签名" name={['sms', 'signName']}>
                      <Input />
                    </Form.Item>
                    <Form.Item label="模板编码" name={['sms', 'templateCode']}>
                      <Input />
                    </Form.Item>
                  </div>
                </TabPane>

                <TabPane
                  key="push"
                  tab={
                    <span>
                      <NotificationOutlined />
                      推送
                    </span>
                  }
                >
                  <div className={styles.formGrid}>
                    <Form.Item label="启用" name={['push', 'enabled']} valuePropName="checked">
                      <Switch />
                    </Form.Item>
                    <Form.Item label="服务商" name={['push', 'provider']}>
                      <Input />
                    </Form.Item>
                    <Form.Item label="Endpoint" name={['push', 'endpoint']}>
                      <Input />
                    </Form.Item>
                    <Form.Item label="AppKey" name={['push', 'appKey']}>
                      <Input autoComplete="off" />
                    </Form.Item>
                    <Form.Item label="AppSecret" name={['push', 'appSecret']}>
                      <Input.Password
                        autoComplete="new-password"
                        placeholder={secretPlaceholder(secretState.pushSecret)}
                      />
                    </Form.Item>
                  </div>
                </TabPane>
              </Tabs>
              <Space>
                <Button type="primary" icon={<SaveOutlined />} loading={saving} onClick={handleSave}>
                  保存
                </Button>
                <Button onClick={() => loadConfig()}>刷新</Button>
              </Space>
            </Space>
          </Form>
        </Spin>
      </Card>
    </PageContainer>
  );
};

export default SystemConfigPage;
