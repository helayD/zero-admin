import {
  AuditOutlined,
  EyeInvisibleOutlined,
  EyeOutlined,
  LockOutlined,
  SafetyCertificateOutlined,
  TeamOutlined,
  ThunderboltOutlined,
  UserOutlined,
} from '@ant-design/icons';
import { LoginForm, ProFormCheckbox, ProFormText } from '@ant-design/pro-form';
import { Alert, Form, message } from 'antd';
import React, { useCallback, useEffect, useRef, useState } from 'react';
import { history, SelectLang, useIntl, useModel } from 'umi';

import { login } from '@/services/ant-design-pro/api';

import styles from './index.less';

const REMEMBER_ACCOUNT_KEY = 'zero_admin_remember_account';

const trustStats = [
  { value: '7+', label: '业务域统一治理' },
  { value: 'RBAC', label: '权限分层管控' },
  { value: 'Audit', label: '操作全程留痕' },
];

const governanceHighlights = [
  {
    icon: <TeamOutlined />,
    title: '统一业务入口',
    description: '系统、会员、商品、订单、营销、内容集中管理。',
  },
  {
    icon: <ThunderboltOutlined />,
    title: '高效运营协同',
    description: '关键流程和经营数据在同一控制台完成闭环。',
  },
  {
    icon: <SafetyCertificateOutlined />,
    title: '企业级安全治理',
    description: '权限、登录、审计日志共同支撑后台访问安全。',
  },
];

const emptyLoginState: API.LoginResult = {
  code: '',
  message: '',
  data: { token: '' },
};

const Login: React.FC = () => {
  const [form] = Form.useForm<API.LoginParams>();
  const [userLoginState, setUserLoginState] = useState<API.LoginResult>(emptyLoginState);
  const [submitting, setSubmitting] = useState(false);
  const [loginAttempts, setLoginAttempts] = useState(0);
  const [capsLockOn, setCapsLockOn] = useState(false);
  const { initialState, setInitialState } = useModel('@@initialState');
  const intl = useIntl();
  const accountInputRef = useRef<any>(null);
  const passwordInputRef = useRef<any>(null);

  useEffect(() => {
    const saved = localStorage.getItem(REMEMBER_ACCOUNT_KEY);
    if (saved) {
      form.setFieldsValue({ account: saved, autoLogin: true });
      setTimeout(() => passwordInputRef.current?.focus(), 100);
      return;
    }

    setTimeout(() => accountInputRef.current?.focus(), 100);
  }, [form]);

  const fetchUserInfo = useCallback(async () => {
    const userInfo = await initialState?.fetchUserInfo?.();
    if (userInfo) {
      await setInitialState((s) => ({ ...s, currentUser: userInfo }));
    }
  }, [initialState, setInitialState]);

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

      <main className={styles.shell}>
        <aside className={styles.brandPanel} aria-label="产品介绍">
          <div className={styles.brandHeader}>
            <img src="/logo.svg" alt="" className={styles.brandLogo} aria-hidden="true" />
            <div>
              <div className={styles.brandName}>九克城</div>
              <div className={styles.brandTagline}>Zero-Admin Enterprise Console</div>
            </div>
          </div>

          <div className={styles.heroBlock}>
            <div className={styles.heroEyebrow}>安全、稳定、清晰的后台入口</div>
            <h1 className={styles.heroTitle}>登录企业管理控制台</h1>
            <p className={styles.heroDescription}>
              为多业务域运营团队提供统一的账号访问、权限治理和审计能力。
            </p>
          </div>

          <div className={styles.statGrid} aria-label="平台能力摘要">
            {trustStats.map((item) => (
              <div key={item.label} className={styles.statCard}>
                <div className={styles.statValue}>{item.value}</div>
                <div className={styles.statLabel}>{item.label}</div>
              </div>
            ))}
          </div>

          <ul className={styles.highlightList}>
            {governanceHighlights.map((item) => (
              <li key={item.title} className={styles.highlightItem}>
                <span className={styles.highlightIcon} aria-hidden="true">
                  {item.icon}
                </span>
                <div>
                  <div className={styles.highlightTitle}>{item.title}</div>
                  <div className={styles.highlightDesc}>{item.description}</div>
                </div>
              </li>
            ))}
          </ul>
        </aside>

        <section className={styles.loginPanel} aria-label="登录入口">
          <div className={styles.panelTopBar}>
            <div className={styles.secureBadge}>
              <AuditOutlined aria-hidden="true" />
              <span>受控访问</span>
            </div>
            <div className={styles.lang}>{SelectLang && <SelectLang />}</div>
          </div>

          <div id="login-form" className={styles.loginCard} role="form" aria-label="登录表单">
            <div className={styles.loginHeader}>
              <h2 className={styles.loginTitle}>欢迎回来</h2>
              <p className={styles.loginSubtitle}>请输入后台账号和密码继续访问</p>
            </div>

            <div className={styles.formNotice}>
              <SafetyCertificateOutlined aria-hidden="true" />
              <span>系统将记录本次登录行为，用于账号安全与合规审计。</span>
            </div>

            {hasLoginError && (
              <Alert
                className={styles.feedbackAlert}
                message={userLoginState.message}
                description={loginAttempts >= 2 ? '请确认账号、密码或联系管理员协助处理。' : undefined}
                type="error"
                showIcon
                closable
                onClose={handleErrorClose}
                role="alert"
                aria-live="assertive"
              />
            )}

            <div className={styles.loginFormWrapper}>
              <LoginForm<API.LoginParams>
                form={form}
                logo={null}
                title=""
                subTitle=""
                initialValues={{ autoLogin: true }}
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
                    disabled: submitting,
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
                    disabled: submitting,
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
                    disabled: submitting,
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
                  <span className={styles.supportText}>忘记密码请联系管理员</span>
                </div>
              </LoginForm>
            </div>

            <div className={styles.meta}>仅限授权人员访问，请勿在公共设备保存账号。</div>
          </div>
        </section>
      </main>
    </div>
  );
};

export default Login;
