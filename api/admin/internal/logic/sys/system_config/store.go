package system_config

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/feihua/zero-admin/api/admin/internal/config"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/zeromicro/go-zero/core/logc"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	groupOSS  = "oss"
	groupSMS  = "sms"
	groupPush = "push"
)

type configRecord struct {
	ID          int64     `gorm:"column:id"`
	ConfigGroup string    `gorm:"column:config_group"`
	ConfigKey   string    `gorm:"column:config_key"`
	ConfigValue string    `gorm:"column:config_value"`
	ValueType   string    `gorm:"column:value_type"`
	IsSecret    int32     `gorm:"column:is_secret"`
	Remark      string    `gorm:"column:remark"`
	CreateBy    string    `gorm:"column:create_by"`
	CreateTime  time.Time `gorm:"column:create_time"`
	UpdateBy    string    `gorm:"column:update_by"`
	UpdateTime  time.Time `gorm:"column:update_time"`
	IsDeleted   int32     `gorm:"column:is_deleted"`
}

func (*configRecord) TableName() string {
	return "sys_system_config"
}

// EffectiveOSSConfig 返回上传接口实际使用的 OSS 配置；后台系统配置优先，运行时 YAML 兜底。
func EffectiveOSSConfig(ctx context.Context, svcCtx *svc.ServiceContext) config.OSSConfig {
	ossConfig := svcCtx.Config.SystemConfig.OSS
	if svcCtx.DB == nil {
		return ossConfig
	}

	values, err := queryGroupValues(ctx, svcCtx.DB, groupOSS)
	if err != nil {
		logc.Errorf(ctx, "读取 OSS 系统配置失败,使用运行时配置兜底,异常:%s", err.Error())
		return ossConfig
	}

	if value := strings.TrimSpace(values["endpoint"]); value != "" {
		ossConfig.Endpoint = value
	}
	if value := strings.TrimSpace(values["accessKeyId"]); value != "" {
		ossConfig.AccessKeyID = value
	}
	if value := strings.TrimSpace(values["accessKeySecret"]); value != "" {
		ossConfig.AccessKeySecret = value
	}
	if value := strings.TrimSpace(values["bucketName"]); value != "" {
		ossConfig.BucketName = value
	}
	if value := strings.TrimSpace(values["url"]); value != "" {
		ossConfig.URL = value
	}
	if value := strings.TrimSpace(values["maxSizeMb"]); value != "" {
		if maxSizeMB, parseErr := strconv.ParseInt(value, 10, 64); parseErr == nil && maxSizeMB > 0 {
			ossConfig.MaxSizeMB = maxSizeMB
		}
	}

	return ossConfig
}

func QuerySystemConfig(ctx context.Context, svcCtx *svc.ServiceContext) (types.SystemConfigData, error) {
	data := defaultSystemConfigData(svcCtx.Config.SystemConfig)
	if svcCtx.DB == nil {
		redactSecretValues(&data)
		return data, nil
	}

	values, err := queryAllValues(ctx, svcCtx.DB)
	if err != nil {
		return data, fmt.Errorf("读取系统配置失败: %w", err)
	}

	applyValues(&data, values)
	redactSecretValues(&data)
	return data, nil
}

func SaveSystemConfig(ctx context.Context, svcCtx *svc.ServiceContext, req *types.SaveSystemConfigReq, operator string) error {
	if svcCtx.DB == nil {
		return errors.New("系统配置数据库未初始化")
	}

	current, err := queryAllValues(ctx, svcCtx.DB)
	if err != nil {
		return fmt.Errorf("读取当前系统配置失败: %w", err)
	}

	records := buildSaveRecords(req, current, operator)
	return svcCtx.DB.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "config_group"},
			{Name: "config_key"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"config_value": gorm.Expr("VALUES(config_value)"),
			"value_type":   gorm.Expr("VALUES(value_type)"),
			"is_secret":    gorm.Expr("VALUES(is_secret)"),
			"remark":       gorm.Expr("VALUES(remark)"),
			"update_by":    gorm.Expr("VALUES(update_by)"),
			"update_time":  gorm.Expr("VALUES(update_time)"),
			"is_deleted":   int32(1),
		}),
	}).Create(&records).Error
}

func queryAllValues(ctx context.Context, db *gorm.DB) (map[string]map[string]string, error) {
	var rows []configRecord
	if err := db.WithContext(ctx).Where("is_deleted = ?", 1).Find(&rows).Error; err != nil {
		return nil, err
	}

	values := make(map[string]map[string]string)
	for _, row := range rows {
		groupValues := values[row.ConfigGroup]
		if groupValues == nil {
			groupValues = make(map[string]string)
			values[row.ConfigGroup] = groupValues
		}
		groupValues[row.ConfigKey] = row.ConfigValue
	}
	return values, nil
}

func queryGroupValues(ctx context.Context, db *gorm.DB, group string) (map[string]string, error) {
	var rows []configRecord
	if err := db.WithContext(ctx).Where("config_group = ? AND is_deleted = ?", group, 1).Find(&rows).Error; err != nil {
		return nil, err
	}

	values := make(map[string]string, len(rows))
	for _, row := range rows {
		values[row.ConfigKey] = row.ConfigValue
	}
	return values, nil
}

func defaultSystemConfigData(c config.SystemConfig) types.SystemConfigData {
	return types.SystemConfigData{
		OSS: types.SystemOSSConfig{
			Endpoint:        c.OSS.Endpoint,
			AccessKeyID:     c.OSS.AccessKeyID,
			AccessKeySecret: c.OSS.AccessKeySecret,
			BucketName:      c.OSS.BucketName,
			URL:             c.OSS.URL,
			MaxSizeMB:       c.OSS.MaxSizeMB,
		},
		SMS: types.SystemSMSConfig{
			Enabled:         c.SMS.Enabled,
			Provider:        c.SMS.Provider,
			Endpoint:        c.SMS.Endpoint,
			AccessKeyID:     c.SMS.AccessKeyID,
			AccessKeySecret: c.SMS.AccessKeySecret,
			SignName:        c.SMS.SignName,
			TemplateCode:    c.SMS.TemplateCode,
		},
		Push: types.SystemPushConfig{
			Enabled:   c.Push.Enabled,
			Provider:  c.Push.Provider,
			Endpoint:  c.Push.Endpoint,
			AppKey:    c.Push.AppKey,
			AppSecret: c.Push.AppSecret,
		},
	}
}

func applyValues(data *types.SystemConfigData, values map[string]map[string]string) {
	if groupValues := values[groupOSS]; groupValues != nil {
		applyString(&data.OSS.Endpoint, groupValues["endpoint"])
		applyString(&data.OSS.AccessKeyID, groupValues["accessKeyId"])
		applyString(&data.OSS.AccessKeySecret, groupValues["accessKeySecret"])
		applyString(&data.OSS.BucketName, groupValues["bucketName"])
		applyString(&data.OSS.URL, groupValues["url"])
		if value := strings.TrimSpace(groupValues["maxSizeMb"]); value != "" {
			if maxSizeMB, err := strconv.ParseInt(value, 10, 64); err == nil {
				data.OSS.MaxSizeMB = maxSizeMB
			}
		}
	}
	if groupValues := values[groupSMS]; groupValues != nil {
		data.SMS.Enabled = parseBool(groupValues["enabled"], data.SMS.Enabled)
		applyString(&data.SMS.Provider, groupValues["provider"])
		applyString(&data.SMS.Endpoint, groupValues["endpoint"])
		applyString(&data.SMS.AccessKeyID, groupValues["accessKeyId"])
		applyString(&data.SMS.AccessKeySecret, groupValues["accessKeySecret"])
		applyString(&data.SMS.SignName, groupValues["signName"])
		applyString(&data.SMS.TemplateCode, groupValues["templateCode"])
	}
	if groupValues := values[groupPush]; groupValues != nil {
		data.Push.Enabled = parseBool(groupValues["enabled"], data.Push.Enabled)
		applyString(&data.Push.Provider, groupValues["provider"])
		applyString(&data.Push.Endpoint, groupValues["endpoint"])
		applyString(&data.Push.AppKey, groupValues["appKey"])
		applyString(&data.Push.AppSecret, groupValues["appSecret"])
	}
}

func buildSaveRecords(req *types.SaveSystemConfigReq, current map[string]map[string]string, operator string) []configRecord {
	now := time.Now()
	records := []configRecord{
		newRecord(groupOSS, "endpoint", req.OSS.Endpoint, "string", 0, "OSS Endpoint", operator, now),
		newRecord(groupOSS, "accessKeyId", req.OSS.AccessKeyID, "string", 1, "OSS AccessKey ID", operator, now),
		newRecord(groupOSS, "bucketName", req.OSS.BucketName, "string", 0, "OSS Bucket", operator, now),
		newRecord(groupOSS, "url", req.OSS.URL, "string", 0, "OSS 公开访问域名", operator, now),
		newRecord(groupOSS, "maxSizeMb", strconv.FormatInt(req.OSS.MaxSizeMB, 10), "int", 0, "上传文件大小上限 MB", operator, now),
		newRecord(groupSMS, "enabled", strconv.FormatBool(req.SMS.Enabled), "bool", 0, "短信配置启用状态", operator, now),
		newRecord(groupSMS, "provider", req.SMS.Provider, "string", 0, "短信服务商", operator, now),
		newRecord(groupSMS, "endpoint", req.SMS.Endpoint, "string", 0, "短信 Endpoint", operator, now),
		newRecord(groupSMS, "accessKeyId", req.SMS.AccessKeyID, "string", 1, "短信 AccessKey ID", operator, now),
		newRecord(groupSMS, "signName", req.SMS.SignName, "string", 0, "短信签名", operator, now),
		newRecord(groupSMS, "templateCode", req.SMS.TemplateCode, "string", 0, "短信模板编码", operator, now),
		newRecord(groupPush, "enabled", strconv.FormatBool(req.Push.Enabled), "bool", 0, "推送配置启用状态", operator, now),
		newRecord(groupPush, "provider", req.Push.Provider, "string", 0, "推送服务商", operator, now),
		newRecord(groupPush, "endpoint", req.Push.Endpoint, "string", 0, "推送 Endpoint", operator, now),
		newRecord(groupPush, "appKey", req.Push.AppKey, "string", 1, "推送 AppKey", operator, now),
	}

	if secret := preserveSecret(req.OSS.AccessKeySecret, current[groupOSS]["accessKeySecret"]); secret != "" {
		records = append(records, newRecord(groupOSS, "accessKeySecret", secret, "string", 1, "OSS AccessKey Secret", operator, now))
	}
	if secret := preserveSecret(req.SMS.AccessKeySecret, current[groupSMS]["accessKeySecret"]); secret != "" {
		records = append(records, newRecord(groupSMS, "accessKeySecret", secret, "string", 1, "短信 AccessKey Secret", operator, now))
	}
	if secret := preserveSecret(req.Push.AppSecret, current[groupPush]["appSecret"]); secret != "" {
		records = append(records, newRecord(groupPush, "appSecret", secret, "string", 1, "推送 AppSecret", operator, now))
	}
	return records
}

func newRecord(group, key, value, valueType string, isSecret int32, remark, operator string, now time.Time) configRecord {
	return configRecord{
		ConfigGroup: group,
		ConfigKey:   key,
		ConfigValue: strings.TrimSpace(value),
		ValueType:   valueType,
		IsSecret:    isSecret,
		Remark:      remark,
		CreateBy:    operator,
		CreateTime:  now,
		UpdateBy:    operator,
		UpdateTime:  now,
		IsDeleted:   1,
	}
}

func applyString(target *string, value string) {
	if value = strings.TrimSpace(value); value != "" {
		*target = value
	}
}

func parseBool(value string, fallback bool) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func preserveSecret(incoming, current string) string {
	incoming = strings.TrimSpace(incoming)
	if incoming != "" {
		return incoming
	}
	return strings.TrimSpace(current)
}

func redactSecretValues(data *types.SystemConfigData) {
	data.OSS.AccessKeySecretConfigured = strings.TrimSpace(data.OSS.AccessKeySecret) != ""
	data.OSS.AccessKeySecret = ""
	data.SMS.AccessKeySecretConfigured = strings.TrimSpace(data.SMS.AccessKeySecret) != ""
	data.SMS.AccessKeySecret = ""
	data.Push.AppSecretConfigured = strings.TrimSpace(data.Push.AppSecret) != ""
	data.Push.AppSecret = ""
}
