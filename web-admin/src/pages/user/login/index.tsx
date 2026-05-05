import {
  BarChartOutlined,
  DashboardOutlined,
  LockOutlined,
  SafetyCertificateOutlined,
  UserOutlined,
} from '@ant-design/icons';
import { LoginForm, ProFormCheckbox, ProFormText } from '@ant-design/pro-form';
import { Alert, message } from 'antd';
import React, { useEffect, useState } from 'react';
import { history, SelectLang, useIntl, useModel } from 'umi';

import { login } from '@/services/ant-design-pro/api';

import styles from './index.less';

const REMEMBER_ACCOUNT_KEY = 'zero_admin_remember_account';

const moduleTags = ['系统治理', '会员中心', '商品管理', '订单履约', '营销运营', '内容与搜索'];

const governanceHighlights = [
  {
    icon: <DashboardOutlined />,
    title: '多业务域协同',
    description: '覆盖系统、会员、商品、订单、营销、内容与搜索等核心后台场景。',
  },
  {
    icon: <BarChartOutlined />,
    title: '经营视图统一',
    description: '数据看板、流程操作与组织协同在一个入口完成闭环联动。',
  },
  {
    icon: <SafetyCertificateOutlined />,
    title: '权限审计可追溯',
    description: '角色授权、登录日志与操作日志共同支撑企业级治理要求。',
  },
];

const trustTags = ['go-zero 微服务架构', 'RBAC 权限模型', '审计链路留痕'];

const emptyLoginState: API.LoginResult = {
  code: '',
  message: '',
  data: {
    token: '',
  },
};

const LoginMessage: React.FC<{
  content: string;
  onClose?: () => void;
}> = ({ content, onClose }) => (
  <Alert
    style={{
      marginBottom: 20,
      borderRadius: 16,
    }}
    message={content}
    type="error"
    showIcon
    closable
    onClose={onClose}
  />
);

const Login: React.FC = () => {
  const [userLoginState, setUserLoginState] = useState<API.LoginResult>(emptyLoginState);
  const [submitting, setSubmitting] = useState(false);
  const [rememberAccount, setRememberAccount] = useState('');
  const { initialState, setInitialState } = useModel('@@initialState');
  const intl = useIntl();

  useEffect(() => {
    const saved = localStorage.getItem(REMEMBER_ACCOUNT_KEY);
    if (saved) {
      setRememberAccount(saved);
    }
  }, []);

  const fetchUserInfo = async () => {
    const userInfo = await initialState?.fetchUserInfo?.();
    if (userInfo) {
      await setInitialState((s) => ({
        ...s,
        currentUser: userInfo,
      }));
    }
  };

  const handleSubmit = async (values: API.LoginParams) => {
    setUserLoginState(emptyLoginState);
    setSubmitting(true);

    try {
      const res = await login(values);
      if (res?.code === '000000') {
        localStorage.setItem('token', res.data.token);

        if (values.autoLogin) {
          localStorage.setItem(REMEMBER_ACCOUNT_KEY, values.account);
        } else {
          localStorage.removeItem(REMEMBER_ACCOUNT_KEY);
        }

        message.success(
          intl.formatMessage({
            id: 'pages.login.success',
            defaultMessage: '登录成功！',
          }),
        );
        await fetchUserInfo();
        if (!history) {
          return false;
        }
        const { query } = history.location;
        const { redirect } = (query || {}) as { redirect?: string };
        history.push(redirect || '/');
        return true;
      }

      setUserLoginState({
        code: res?.code || '111111',
        message:
          res?.message ||
          intl.formatMessage({
            id: 'pages.login.accountLogin.errorMessage',
            defaultMessage: '账号或密码错误，请重新输入',
          }),
        data: {
          token: '',
        },
      });
      return false;
    } catch (error) {
      const fallbackMessage = intl.formatMessage({
        id: 'pages.login.failure',
        defaultMessage: '登录失败，请重试！',
      });
      setUserLoginState({
        code: 'NETWORK_ERROR',
        message: fallbackMessage,
        data: {
          token: '',
        },
      });
      message.error(fallbackMessage);
      return false;
    } finally {
      setSubmitting(false);
    }
  };

  const hasLoginError =
    userLoginState.code !== '' &&
    userLoginState.code !== '000000' &&
    Boolean(userLoginState.message);

  return (
    <div className={styles.container} role="main" aria-label="登录页面">
      <div className={styles.brandSide} aria-hidden="true">
        <div className={styles.brandContent}>
          <div className={styles.brandEyebrow}>ZERO-ADMIN ENTERPRISE CONSOLE</div>

          <div className={styles.brandHero}>
            <img src="/logo.svg" alt="" className={styles.brandLogo} aria-hidden="true" />
            <div className={styles.brandHeroText}>
              <div className={styles.brandTitle}>九克城</div>
              <div className={styles.brandTagline}>企业级后台管理控制台</div>
            </div>
          </div>

          <div className={styles.brandHeadline}>一站式企业后台管理</div>

          <div className={styles.brandSubtitle}>
            基于 go-zero 微服务架构，聚合系统、会员、商品、订单、营销、内容核心模块。
          </div>

          <div className={styles.modulePills}>
            {moduleTags.map((item) => (
              <span key={item} className={styles.modulePill}>
                {item}
              </span>
            ))}
          </div>

          <div className={styles.brandGrid}>
            {governanceHighlights.map((item) => (
              <div key={item.title} className={styles.brandPanel}>
                <div className={styles.brandPanelIcon}>{item.icon}</div>
                <div className={styles.brandPanelTitle}>{item.title}</div>
                <div className={styles.brandPanelDesc}>{item.description}</div>
              </div>
            ))}
          </div>

          <div className={styles.trustStrip}>
            {trustTags.map((item) => (
              <span key={item} className={styles.trustTag}>
                {item}
              </span>
            ))}
          </div>
        </div>
      </div>

      <div className={styles.loginSide}>
        <div className={styles.loginShell}>
          <div className={styles.lang} data-lang>
            {SelectLang && <SelectLang />}
          </div>

          <div className={styles.loginCard} role="form" aria-label="登录表单">
            <div className={styles.loginBadge}>受控访问入口</div>

            <div className={styles.loginHeader}>
              <div className={styles.loginLogo}>
                <img src="/logo.svg" alt="" aria-hidden="true" />
                <span>九克城后台</span>
              </div>
              <div className={styles.loginTitle}>登录管理控制台</div>
              <div className={styles.loginSubtitle}>
                使用已分配的后台账号进入系统。
              </div>
            </div>

            <Alert
              showIcon
              type="info"
              className={styles.securityAlert}
              message="登录前提示"
              description="若账号未开通或密码需要重置，请联系管理员处理。"
            />

            {hasLoginError && (
              <LoginMessage
                content={userLoginState.message}
                onClose={() => setUserLoginState(emptyLoginState)}
              />
            )}

            <div className={styles.loginFormWrapper}>
              <LoginForm<API.LoginParams>
                logo={null}
                title=""
                subTitle=""
                initialValues={{
                  autoLogin: true,
                  account: rememberAccount,
                }}
                submitter={{
                  searchConfig: {
                    submitText: submitting
                      ? intl.formatMessage({
                          id: 'pages.login.submitting',
                          defaultMessage: '登录中...',
                        })
                      : intl.formatMessage({
                          id: 'pages.login.submit',
                          defaultMessage: '进入控制台',
                        }),
                  },
                  submitButtonProps: {
                    size: 'large',
                    className: styles.submitButton,
                    loading: submitting,
                    disabled: submitting,
                  },
                }}
                onFinish={async (values) => handleSubmit(values)}
              >
                <ProFormText
                  label="账号"
                  name="account"
                  fieldProps={{
                    size: 'large',
                    prefix: <UserOutlined className={styles.prefixIcon} aria-hidden="true" />,
                    autoComplete: 'username',
                    maxLength: 64,
                    allowClear: true,
                    disabled: submitting,
                    autoFocus: !rememberAccount,
                  }}
                  placeholder={intl.formatMessage({
                    id: 'pages.login.username.placeholder',
                    defaultMessage: '请输入用户名或手机号',
                  })}
                  rules={[
                    {
                      required: true,
                      message: intl.formatMessage({
                        id: 'pages.login.username.required',
                        defaultMessage: '请输入用户名！',
                      }),
                    },
                  ]}
                />

                <ProFormText.Password
                  label="密码"
                  name="password"
                  fieldProps={{
                    size: 'large',
                    prefix: <LockOutlined className={styles.prefixIcon} aria-hidden="true" />,
                    autoComplete: 'current-password',
                    maxLength: 64,
                    disabled: submitting,
                    autoFocus: !!rememberAccount,
                  }}
                  placeholder={intl.formatMessage({
                    id: 'pages.login.password.placeholder',
                    defaultMessage: '请输入登录密码',
                  })}
                  rules={[
                    {
                      required: true,
                      message: intl.formatMessage({
                        id: 'pages.login.password.required',
                        defaultMessage: '请输入密码！',
                      }),
                    },
                  ]}
                />

                <div className={styles.formOptions}>
                  <ProFormCheckbox noStyle name="autoLogin">
                    {intl.formatMessage({
                      id: 'pages.login.rememberMe',
                      defaultMessage: '记住账号',
                    })}
                  </ProFormCheckbox>
                  <span className={styles.supportText}>
                    {intl.formatMessage({
                      id: 'pages.login.forgotPassword',
                      defaultMessage: '忘记密码请联系平台管理员',
                    })}
                  </span>
                </div>

                <div className={styles.submitHint}>
                  登录即表示同意后台访问控制策略。
                </div>
              </LoginForm>
            </div>
          </div>

          <div className={styles.footer}>
            &copy; {new Date().getFullYear()} 九克城 · Zero-Admin Enterprise Console
          </div>
        </div>
      </div>
    </div>
  );
};

export default Login;
