import {
  EyeInvisibleOutlined,
  EyeOutlined,
  LockOutlined,
  SafetyCertificateOutlined,
  TeamOutlined,
  ThunderboltOutlined,
  UserOutlined,
} from '@ant-design/icons';
import { LoginForm, ProFormCheckbox, ProFormText } from '@ant-design/pro-form';
import { Alert, message } from 'antd';
import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { history, SelectLang, useIntl, useModel } from 'umi';

import { login } from '@/services/ant-design-pro/api';

import styles from './index.less';

const REMEMBER_ACCOUNT_KEY = 'zero_admin_remember_account';
const MAX_LOGIN_ATTEMPTS = 5;

const governanceHighlights = [
  {
    icon: <TeamOutlined />,
    title: '全业务域聚合',
    description: '系统、会员、商品、订单、营销、内容一个入口闭环。',
  },
  {
    icon: <ThunderboltOutlined />,
    title: '实时数据驱动',
    description: '经营看板与流程操作共享同一数据源，决策更高效。',
  },
  {
    icon: <SafetyCertificateOutlined />,
    title: '审计留痕',
    description: 'RBAC 权限、登录与操作日志支撑企业级治理。',
  },
];

const emptyLoginState: API.LoginResult = {
  code: '',
  message: '',
  data: { token: '' },
};

const Login: React.FC = () => {
  const [userLoginState, setUserLoginState] = useState<API.LoginResult>(emptyLoginState);
  const [submitting, setSubmitting] = useState(false);
  const [rememberAccount, setRememberAccount] = useState('');
  const [loginAttempts, setLoginAttempts] = useState(0);
  const [capsLockOn, setCapsLockOn] = useState(false);
  const { initialState, setInitialState } = useModel('@@initialState');
  const intl = useIntl();
  const accountInputRef = useRef<any>(null);
  const passwordInputRef = useRef<any>(null);

  useEffect(() => {
    const saved = localStorage.getItem(REMEMBER_ACCOUNT_KEY);
    if (saved) {
      setRememberAccount(saved);
      setTimeout(() => passwordInputRef.current?.focus(), 100);
    } else {
      setTimeout(() => accountInputRef.current?.focus(), 100);
    }
  }, []);

  const fetchUserInfo = async () => {
    const userInfo = await initialState?.fetchUserInfo?.();
    if (userInfo) {
      await setInitialState((s) => ({ ...s, currentUser: userInfo }));
    }
  };

  const handleSubmit = useCallback(
    async (values: API.LoginParams) => {
      setUserLoginState(emptyLoginState);
      setSubmitting(true);

      try {
        const res = await login(values);
        if (res?.code === '000000') {
          localStorage.setItem('token', res.data.token);
          if (values.autoLogin && values.account) {
            localStorage.setItem(REMEMBER_ACCOUNT_KEY, values.account);
          } else {
            localStorage.removeItem(REMEMBER_ACCOUNT_KEY);
          }

          message.success(
            intl.formatMessage({
              id: 'pages.login.success',
              defaultMessage: '登录成功',
            }),
          );
          setLoginAttempts(0);

          await fetchUserInfo();
          if (!history) {
            return false;
          }
          const { query } = history.location;
          const { redirect } = (query || {}) as { redirect?: string };
          history.push(redirect || '/');
          return true;
        }

        setLoginAttempts((prev) => prev + 1);
        setUserLoginState({
          code: res?.code || '111111',
          message:
            res?.message ||
            intl.formatMessage({
              id: 'pages.login.accountLogin.errorMessage',
              defaultMessage: '账号或密码错误，请重新输入',
            }),
          data: { token: '' },
        });

        setTimeout(() => {
          passwordInputRef.current?.focus();
          passwordInputRef.current?.select?.();
        }, 100);

        return false;
      } catch (error) {
        setLoginAttempts((prev) => prev + 1);
        const fallbackMessage = intl.formatMessage({
          id: 'pages.login.failure',
          defaultMessage: '登录失败，请稍后再试',
        });
        setUserLoginState({
          code: 'NETWORK_ERROR',
          message: fallbackMessage,
          data: { token: '' },
        });
        message.error(fallbackMessage);
        return false;
      } finally {
        setSubmitting(false);
      }
    },
    [intl, fetchUserInfo],
  );

  const hasLoginError =
    userLoginState.code !== '' &&
    userLoginState.code !== '000000' &&
    Boolean(userLoginState.message);

  const isAccountLocked = useMemo(() => loginAttempts >= MAX_LOGIN_ATTEMPTS, [loginAttempts]);

  const handleErrorClose = useCallback(() => {
    setUserLoginState(emptyLoginState);
  }, []);

  const detectCapsLock = useCallback((e: React.KeyboardEvent<HTMLInputElement>) => {
    if (typeof e.getModifierState === 'function') {
      setCapsLockOn(e.getModifierState('CapsLock'));
    }
  }, []);

  return (
    <div className={styles.container} role="main" aria-label="登录页面">
      <a href="#login-form" className={styles.skipLink}>
        跳转到登录表单
      </a>

      <aside className={styles.brandSide} aria-label="产品介绍">
        <div className={styles.brandContent}>
          <div className={styles.brandHero}>
            <img src="/logo.svg" alt="" className={styles.brandLogo} aria-hidden="true" />
            <div>
              <div className={styles.brandTitle}>九克城</div>
              <div className={styles.brandTagline}>Enterprise Console</div>
            </div>
          </div>

          <h1 className={styles.brandHeadline}>企业级后台管理</h1>
          <p className={styles.brandSubtitle}>
            基于 go-zero 微服务架构，聚合业务、运营、数据与治理能力。
          </p>

          <ul className={styles.brandHighlights}>
            {governanceHighlights.map((item) => (
              <li key={item.title} className={styles.highlightItem}>
                <span className={styles.highlightIcon} aria-hidden="true">
                  {item.icon}
                </span>
                <div className={styles.highlightText}>
                  <div className={styles.highlightTitle}>{item.title}</div>
                  <div className={styles.highlightDesc}>{item.description}</div>
                </div>
              </li>
            ))}
          </ul>
        </div>

        <div className={styles.brandFooter}>
          <span>RBAC 权限</span>
          <span className={styles.brandFooterDivider} aria-hidden="true" />
          <span>审计链路</span>
          <span className={styles.brandFooterDivider} aria-hidden="true" />
          <span>多业务域</span>
        </div>
      </aside>

      <section className={styles.loginSide} aria-label="登录入口">
        <div className={styles.loginShell}>
          <div className={styles.topBar}>
            <div className={styles.brandCompact}>
              <img src="/logo.svg" alt="" aria-hidden="true" />
              <span>九克城后台</span>
            </div>
            <div className={styles.lang}>{SelectLang && <SelectLang />}</div>
          </div>

          <div
            id="login-form"
            className={styles.loginCard}
            role="form"
            aria-label="登录表单"
          >
            <div className={styles.loginHeader}>
              <h2 className={styles.loginTitle}>登录控制台</h2>
              <p className={styles.loginSubtitle}>请使用已分配的后台账号进入系统</p>
            </div>

            {hasLoginError && (
              <Alert
                className={styles.feedbackAlert}
                message={userLoginState.message}
                type="error"
                showIcon
                closable
                onClose={handleErrorClose}
                role="alert"
                aria-live="assertive"
              />
            )}

            {isAccountLocked && (
              <Alert
                className={styles.feedbackAlert}
                message="账号已被临时锁定"
                description="多次登录失败，请稍后再试或联系管理员"
                type="warning"
                showIcon
                role="alert"
                aria-live="polite"
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
                          defaultMessage: '登录中',
                        })
                      : intl.formatMessage({
                          id: 'pages.login.submit',
                          defaultMessage: '登录',
                        }),
                  },
                  submitButtonProps: {
                    size: 'large',
                    className: styles.submitButton,
                    loading: submitting,
                    disabled: submitting || isAccountLocked,
                    'aria-label': submitting ? '登录中' : '登录',
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
                    disabled: submitting || isAccountLocked,
                    ref: accountInputRef,
                    'aria-label': '账号',
                    'aria-required': 'true',
                    'aria-invalid': hasLoginError ? 'true' : 'false',
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
                        defaultMessage: '请输入用户名',
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
                    iconRender: (visible) =>
                      visible ? (
                        <EyeOutlined aria-label="隐藏密码" role="button" tabIndex={0} />
                      ) : (
                        <EyeInvisibleOutlined aria-label="显示密码" role="button" tabIndex={0} />
                      ),
                    autoComplete: 'current-password',
                    maxLength: 64,
                    disabled: submitting || isAccountLocked,
                    ref: passwordInputRef,
                    onKeyDown: detectCapsLock,
                    onKeyUp: detectCapsLock,
                    'aria-label': '密码',
                    'aria-required': 'true',
                    'aria-invalid': hasLoginError ? 'true' : 'false',
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
                        defaultMessage: '请输入密码',
                      }),
                    },
                  ]}
                />

                {capsLockOn && (
                  <div className={styles.capsLockHint} role="status" aria-live="polite">
                    大写锁定已开启
                  </div>
                )}

                <div className={styles.formOptions}>
                  <ProFormCheckbox noStyle name="autoLogin">
                    {intl.formatMessage({
                      id: 'pages.login.rememberMe',
                      defaultMessage: '记住账号',
                    })}
                  </ProFormCheckbox>
                  <span className={styles.supportText}>
                    忘记密码请联系管理员
                  </span>
                </div>
              </LoginForm>
            </div>

            <div className={styles.meta}>
              登录即表示同意后台访问控制策略
            </div>
          </div>

          <footer className={styles.footer}>
            &copy; {new Date().getFullYear()} 九克城 · Zero-Admin Enterprise Console
          </footer>
        </div>
      </section>
    </div>
  );
};

export default Login;
