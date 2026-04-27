package types

type QuerySystemConfigResp struct {
	Code    string           `json:"code"`
	Message string           `json:"message"`
	Data    SystemConfigData `json:"data"`
}

type SaveSystemConfigReq struct {
	OSS  SystemOSSConfig  `json:"oss"`
	SMS  SystemSMSConfig  `json:"sms,optional"`
	Push SystemPushConfig `json:"push,optional"`
}

type SaveSystemConfigResp struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type SystemConfigData struct {
	OSS  SystemOSSConfig  `json:"oss"`
	SMS  SystemSMSConfig  `json:"sms"`
	Push SystemPushConfig `json:"push"`
}

type SystemOSSConfig struct {
	Endpoint                  string `json:"endpoint,optional"`
	AccessKeyID               string `json:"accessKeyId,optional"`
	AccessKeySecret           string `json:"accessKeySecret,optional"`
	AccessKeySecretConfigured bool   `json:"accessKeySecretConfigured,optional"`
	BucketName                string `json:"bucketName,optional"`
	URL                       string `json:"url,optional"`
	MaxSizeMB                 int64  `json:"maxSizeMb,optional"`
}

type SystemSMSConfig struct {
	Enabled                   bool   `json:"enabled,optional"`
	Provider                  string `json:"provider,optional"`
	Endpoint                  string `json:"endpoint,optional"`
	AccessKeyID               string `json:"accessKeyId,optional"`
	AccessKeySecret           string `json:"accessKeySecret,optional"`
	AccessKeySecretConfigured bool   `json:"accessKeySecretConfigured,optional"`
	SignName                  string `json:"signName,optional"`
	TemplateCode              string `json:"templateCode,optional"`
}

type SystemPushConfig struct {
	Enabled             bool   `json:"enabled,optional"`
	Provider            string `json:"provider,optional"`
	Endpoint            string `json:"endpoint,optional"`
	AppKey              string `json:"appKey,optional"`
	AppSecret           string `json:"appSecret,optional"`
	AppSecretConfigured bool   `json:"appSecretConfigured,optional"`
}
