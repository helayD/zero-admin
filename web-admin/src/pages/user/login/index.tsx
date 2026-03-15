import {
  LockOutlined,
  MobileOutlined,
  UserOutlined,
  SafetyCertificateOutlined,
  DashboardOutlined,
  BarChartOutlined,
} from '@ant-design/icons';
import {Alert, message, Tabs} from 'antd';
import React, {useState} from 'react';
import {ProFormCaptcha, ProFormCheckbox, ProFormText, LoginForm} from '@ant-design/pro-form';
import {useIntl, history, FormattedMessage, SelectLang, useModel} from 'umi';
import {login} from '@/services/ant-design-pro/api';

import styles from './index.less';

const LoginMessage: React.FC<{
  content: string;
}> = ({content}) => (
  <Alert
    style={{
      marginBottom: 24,
      borderRadius: 8,
    }}
    message={content}
    type="error"
    showIcon
  />
);

const Login: React.FC = () => {
  const [userLoginState, setUserLoginState] = useState<API.LoginResult>({
    code: "", message: "", data: {token: ""}
  });
  const [type, setType] = useState<string>('account');
  const {initialState, setInitialState} = useModel('@@initialState');

  const intl = useIntl();

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
    try {
      // 登录
      const res = await login({...values, type});
      if (res.code === '000000') {
        localStorage.setItem("token", res.data.token);
        const defaultLoginSuccessMessage = intl.formatMessage({
          id: 'pages.login.success',
          defaultMessage: '登录成功！',
        });
        message.success(defaultLoginSuccessMessage);
        await fetchUserInfo();
        /** 此方法会跳转到 redirect 参数所在的位置 */
        if (!history) return;
        const {query} = history.location;
        const {redirect} = query as { redirect: string };
        history.push(redirect || '/');
        return;
      }
      // 如果失败去设置用户错误信息
      setUserLoginState(res);
    } catch (error) {
      const defaultLoginFailureMessage = intl.formatMessage({
        id: 'pages.login.failure',
        defaultMessage: '登录失败，请重试！',
      });
      message.error(defaultLoginFailureMessage);
    }
  };
  const {code} = userLoginState;

  return (
    <div className={styles.container}>
      {/* 左侧品牌展示区 */}
      <div className={styles.brandSide}>
        <div className={styles.brandContent}>
          <img src="/logo.svg" alt="logo" className={styles.brandLogo} />
          <div className={styles.brandTitle}>九克城</div>
          <div className={styles.brandSubtitle}>
            企业级一站式后台管理系统<br/>
            高效、安全、智能的业务管理平台
          </div>
          <div className={styles.featureList}>
            <div className={styles.featureItem}>
              <div className={`${styles.featureIcon} ${styles.featureIconBlue}`}>
                <DashboardOutlined />
              </div>
              <span className={styles.featureText}>全方位数据看板，实时监控业务状态</span>
            </div>
            <div className={styles.featureItem}>
              <div className={`${styles.featureIcon} ${styles.featureIconPurple}`}>
                <BarChartOutlined />
              </div>
              <span className={styles.featureText}>多维度数据分析，助力精准决策</span>
            </div>
            <div className={styles.featureItem}>
              <div className={`${styles.featureIcon} ${styles.featureIconGreen}`}>
                <SafetyCertificateOutlined />
              </div>
              <span className={styles.featureText}>细粒度权限管控，保障数据安全</span>
            </div>
          </div>
        </div>
      </div>

      {/* 右侧登录区 */}
      <div className={styles.loginSide}>
        <div className={styles.lang} data-lang>
          {SelectLang && <SelectLang/>}
        </div>

        <div className={styles.loginCard}>
          <div className={styles.loginHeader}>
            <div className={styles.loginLogo}>
              <img src="/logo.svg" alt="logo" />
              <span>九克城</span>
            </div>
            <div className={styles.loginTitle}>欢迎回来</div>
            <div className={styles.loginSubtitle}>登录您的账户以继续使用管理系统</div>
          </div>

          <div className={styles.loginFormWrapper}>
            <LoginForm
              logo={null}
              title=""
              subTitle=""
              initialValues={{
                autoLogin: true,
              }}
              onFinish={async (values) => {
                await handleSubmit(values as API.LoginParams);
              }}
            >
              <Tabs activeKey={type} onChange={setType}>
                <Tabs.TabPane
                  key="account"
                  tab={intl.formatMessage({
                    id: 'pages.login.accountLogin.tab',
                    defaultMessage: '账户密码登录',
                  })}
                />
                <Tabs.TabPane
                  key="mobile"
                  tab={intl.formatMessage({
                    id: 'pages.login.phoneLogin.tab',
                    defaultMessage: '手机号登录',
                  })}
                />
              </Tabs>

              {code === '111111' && type === 'account' && (
                <LoginMessage
                  content={intl.formatMessage({
                    id: 'pages.login.accountLogin.errorMessage',
                    defaultMessage: '账户或密码错误',
                  })}
                />
              )}
              {type === 'account' && (
                <>
                  <ProFormText
                    name="account"
                    fieldProps={{
                      size: 'large',
                      prefix: <UserOutlined className={styles.prefixIcon}/>,
                    }}
                    placeholder={intl.formatMessage({
                      id: 'pages.login.username.placeholder',
                      defaultMessage: '请输入用户名',
                    })}
                    rules={[
                      {
                        required: true,
                        message: (
                          <FormattedMessage
                            id="pages.login.username.required"
                            defaultMessage="请输入用户名!"
                          />
                        ),
                      },
                    ]}
                  />
                  <ProFormText.Password
                    name="password"
                    fieldProps={{
                      size: 'large',
                      prefix: <LockOutlined className={styles.prefixIcon}/>,
                    }}
                    placeholder={intl.formatMessage({
                      id: 'pages.login.password.placeholder',
                      defaultMessage: '请输入密码',
                    })}
                    rules={[
                      {
                        required: true,
                        message: (
                          <FormattedMessage
                            id="pages.login.password.required"
                            defaultMessage="请输入密码！"
                          />
                        ),
                      },
                    ]}
                  />
                </>
              )}

              {code === '111111' && type === 'mobile' && <LoginMessage content="验证码错误"/>}
              {type === 'mobile' && (
                <>
                  <ProFormText
                    fieldProps={{
                      size: 'large',
                      prefix: <MobileOutlined className={styles.prefixIcon}/>,
                    }}
                    name="mobile"
                    placeholder={intl.formatMessage({
                      id: 'pages.login.phoneNumber.placeholder',
                      defaultMessage: '手机号',
                    })}
                    rules={[
                      {
                        required: true,
                        message: (
                          <FormattedMessage
                            id="pages.login.phoneNumber.required"
                            defaultMessage="请输入手机号！"
                          />
                        ),
                      },
                      {
                        pattern: /^1\d{10}$/,
                        message: (
                          <FormattedMessage
                            id="pages.login.phoneNumber.invalid"
                            defaultMessage="手机号格式错误！"
                          />
                        ),
                      },
                    ]}
                  />
                  <ProFormCaptcha
                    fieldProps={{
                      size: 'large',
                      prefix: <LockOutlined className={styles.prefixIcon}/>,
                    }}
                    captchaProps={{
                      size: 'large',
                    }}
                    placeholder={intl.formatMessage({
                      id: 'pages.login.captcha.placeholder',
                      defaultMessage: '请输入验证码',
                    })}
                    captchaTextRender={(timing, count) => {
                      if (timing) {
                        return `${count} ${intl.formatMessage({
                          id: 'pages.getCaptchaSecondText',
                          defaultMessage: '获取验证码',
                        })}`;
                      }
                      return intl.formatMessage({
                        id: 'pages.login.phoneLogin.getVerificationCode',
                        defaultMessage: '获取验证码',
                      });
                    }}
                    name="captcha"
                    rules={[
                      {
                        required: true,
                        message: (
                          <FormattedMessage
                            id="pages.login.captcha.required"
                            defaultMessage="请输入验证码！"
                          />
                        ),
                      },
                    ]}
                    onGetCaptcha={async (phone) => {

                      message.success('获取验证码成功！验证码为：1234');
                    }}
                  />
                </>
              )}
              <div
                style={{
                  marginBottom: 24,
                }}
              >
                <ProFormCheckbox noStyle name="autoLogin">
                  <FormattedMessage id="pages.login.rememberMe" defaultMessage="自动登录"/>
                </ProFormCheckbox>
                <a
                  style={{
                    float: 'right',
                  }}
                >
                  <FormattedMessage id="pages.login.forgotPassword" defaultMessage="忘记密码"/>
                </a>
              </div>
            </LoginForm>
          </div>
        </div>

        <div className={styles.footer}>
          &copy; {new Date().getFullYear()} 九克城 · 企业级管理系统
        </div>
      </div>
    </div>
  );
};

export default Login;
