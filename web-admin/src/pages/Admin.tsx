import React from 'react';
import { SafetyCertificateOutlined } from '@ant-design/icons';
import { Card, Typography } from 'antd';
import { PageHeaderWrapper } from '@ant-design/pro-layout';

const Admin: React.FC = () => {
  return (
    <PageHeaderWrapper
      content="此页面仅管理员可查看"
    >
      <Card>
        <Typography.Title level={3} style={{ textAlign: 'center', color: '#1a1a2e' }}>
          <SafetyCertificateOutlined style={{ marginRight: 8 }} />
          九克城管理后台
        </Typography.Title>
        <Typography.Paragraph style={{ textAlign: 'center', color: '#8c8c8c' }}>
          欢迎使用九克城企业级管理系统，请通过左侧菜单导航至相应功能模块。
        </Typography.Paragraph>
      </Card>
    </PageHeaderWrapper>
  );
};

export default Admin;
