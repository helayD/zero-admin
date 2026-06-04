package sms

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"

	openapiutil "github.com/alibabacloud-go/darabonba-openapi/v2/utils"
	dysmsapi "github.com/alibabacloud-go/dysmsapi-20170525/v5/client"
	"github.com/alibabacloud-go/tea/tea"
)

const (
	AliyunProviderCode = "aliyun"

	aliyunDefaultEndpoint       = "dysmsapi.aliyuncs.com"
	aliyunDefaultAccessKeyIDEnv = "ALIYUN_SMS_ACCESS_KEY_ID"
	aliyunDefaultSecretEnv      = "ALIYUN_SMS_ACCESS_KEY_SECRET"
)

type aliyunSendSmsAPI func(ctx context.Context, runtimeCfg aliyunRuntimeConfig, req *dysmsapi.SendSmsRequest) (*dysmsapi.SendSmsResponse, error)

var aliyunSendSms aliyunSendSmsAPI = sendAliyunSms

// AliyunProvider 阿里云短信网关实现。
//
// CredentialRef 支持:
//   - ref:env:ALIYUN_SMS 读取 ALIYUN_SMS_ACCESS_KEY_ID / ALIYUN_SMS_ACCESS_KEY_SECRET
//   - ref:env:ALIYUN_SMS_ACCESS_KEY_ID,ALIYUN_SMS_ACCESS_KEY_SECRET 分别指定 ID/Secret 环境变量
//   - 空值时兜底读取 ALIYUN_SMS_ACCESS_KEY_ID / ALIYUN_SMS_ACCESS_KEY_SECRET
type AliyunProvider struct{}

func (AliyunProvider) Code() string { return AliyunProviderCode }

func (AliyunProvider) Send(ctx context.Context, mobile string, scene Scene, cfg Config) (*SendResult, error) {
	signName := strings.TrimSpace(cfg.SignName)
	if signName == "" {
		return nil, errors.New("aliyun: signName 不能为空")
	}
	templateCode := strings.TrimSpace(cfg.TemplateCode)
	if templateCode == "" {
		return nil, errors.New("aliyun: templateCode 不能为空")
	}

	accessKeyID, accessKeySecret, err := resolveAliyunCredential(cfg)
	if err != nil {
		return nil, err
	}

	code, err := generateSixDigitCode()
	if err != nil {
		return nil, fmt.Errorf("aliyun: 生成验证码失败: %w", err)
	}

	templateParam, err := json.Marshal(map[string]string{"code": code})
	if err != nil {
		return nil, fmt.Errorf("aliyun: 生成模板参数失败: %w", err)
	}

	timeoutSeconds := cfg.TimeoutSeconds
	if timeoutSeconds <= 0 {
		timeoutSeconds = 5
	}

	endpoint := strings.TrimSpace(cfg.Endpoint)
	if endpoint == "" {
		endpoint = aliyunDefaultEndpoint
	}

	resp, err := aliyunSendSms(ctx, aliyunRuntimeConfig{
		AccessKeyID:     accessKeyID,
		AccessKeySecret: accessKeySecret,
		Endpoint:        endpoint,
		TimeoutSeconds:  timeoutSeconds,
	}, &dysmsapi.SendSmsRequest{
		PhoneNumbers:  tea.String(strings.TrimSpace(mobile)),
		SignName:      tea.String(signName),
		TemplateCode:  tea.String(templateCode),
		TemplateParam: tea.String(string(templateParam)),
	})
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Body == nil {
		return nil, errors.New("aliyun: 返回空响应")
	}
	if got := tea.StringValue(resp.Body.Code); got != "OK" {
		message := strings.TrimSpace(tea.StringValue(resp.Body.Message))
		if message == "" {
			message = "未知错误"
		}
		return nil, fmt.Errorf("aliyun: SendSms 返回失败 code=%s message=%s", got, message)
	}

	expireSeconds := cfg.ExpireSeconds
	if expireSeconds <= 0 {
		expireSeconds = defaultExpireSeconds
	}

	_ = scene
	return &SendResult{
		Code:          code,
		ExpireSeconds: expireSeconds,
		ProviderCode:  AliyunProviderCode,
	}, nil
}

type aliyunRuntimeConfig struct {
	AccessKeyID     string
	AccessKeySecret string
	Endpoint        string
	TimeoutSeconds  int
}

func sendAliyunSms(ctx context.Context, runtimeCfg aliyunRuntimeConfig, req *dysmsapi.SendSmsRequest) (*dysmsapi.SendSmsResponse, error) {
	endpoint := strings.TrimSpace(runtimeCfg.Endpoint)
	if endpoint == "" {
		endpoint = aliyunDefaultEndpoint
	}
	timeoutSeconds := runtimeCfg.TimeoutSeconds
	if timeoutSeconds <= 0 {
		timeoutSeconds = 5
	}
	client, err := dysmsapi.NewClient(&openapiutil.Config{
		AccessKeyId:     tea.String(runtimeCfg.AccessKeyID),
		AccessKeySecret: tea.String(runtimeCfg.AccessKeySecret),
		Endpoint:        tea.String(endpoint),
		ReadTimeout:     tea.Int(timeoutSeconds * 1000),
		ConnectTimeout:  tea.Int(timeoutSeconds * 1000),
	})
	if err != nil {
		return nil, fmt.Errorf("aliyun: 初始化客户端失败: %w", err)
	}

	if ctx == nil {
		return client.SendSms(req)
	}

	type response struct {
		resp *dysmsapi.SendSmsResponse
		err  error
	}
	done := make(chan response, 1)
	go func() {
		resp, err := client.SendSms(req)
		done <- response{resp: resp, err: err}
	}()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("aliyun: 请求已取消: %w", ctx.Err())
	case result := <-done:
		if result.err != nil {
			return nil, fmt.Errorf("aliyun: SendSms 调用失败: %w", result.err)
		}
		return result.resp, nil
	}
}

func resolveAliyunCredential(cfg Config) (string, string, error) {
	accessKeyID := strings.TrimSpace(cfg.AccessKeyID)
	accessKeySecret := strings.TrimSpace(cfg.AccessKeySecret)
	if accessKeyID != "" && accessKeySecret != "" {
		return accessKeyID, accessKeySecret, nil
	}

	idEnv, secretEnv, err := aliyunCredentialEnvNames(cfg.CredentialRef)
	if err != nil {
		return "", "", err
	}

	accessKeyID = strings.TrimSpace(os.Getenv(idEnv))
	accessKeySecret = strings.TrimSpace(os.Getenv(secretEnv))
	if accessKeyID == "" || accessKeySecret == "" {
		return "", "", fmt.Errorf("aliyun: 环境变量[%s/%s]未配置", idEnv, secretEnv)
	}
	return accessKeyID, accessKeySecret, nil
}

func aliyunCredentialEnvNames(ref string) (string, string, error) {
	trimmed := strings.TrimSpace(ref)
	if trimmed == "" {
		return aliyunDefaultAccessKeyIDEnv, aliyunDefaultSecretEnv, nil
	}
	if !strings.HasPrefix(trimmed, "ref:env:") {
		return "", "", errors.New("aliyun: credentialRef 仅支持 ref:env: 前缀")
	}

	spec := strings.TrimSpace(strings.TrimPrefix(trimmed, "ref:env:"))
	if spec == "" {
		return "", "", errors.New("aliyun: credentialRef 环境变量名不能为空")
	}

	parts := strings.Split(spec, ",")
	if len(parts) == 1 {
		prefix := strings.TrimSpace(parts[0])
		return prefix + "_ACCESS_KEY_ID", prefix + "_ACCESS_KEY_SECRET", nil
	}
	if len(parts) == 2 {
		idEnv := strings.TrimSpace(parts[0])
		secretEnv := strings.TrimSpace(parts[1])
		if idEnv == "" || secretEnv == "" {
			return "", "", errors.New("aliyun: credentialRef 环境变量名不能为空")
		}
		return idEnv, secretEnv, nil
	}
	return "", "", errors.New("aliyun: credentialRef 格式非法")
}

func generateSixDigitCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func init() {
	RegisterProvider(AliyunProvider{})
}
