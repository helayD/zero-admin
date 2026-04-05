package types

import "encoding/json"

type CreateChannelIntegrationTemplateReq struct {
	TemplateCode         string          `json:"templateCode"`
	TemplateName         string          `json:"templateName"`
	TemplateType         string          `json:"templateType"`
	TargetCode           string          `json:"targetCode"`
	ScopeType            string          `json:"scopeType"`
	PlatformId           int64           `json:"platformId,optional"`
	TenantId             int64           `json:"tenantId,optional"`
	MerchantId           int64           `json:"merchantId,optional"`
	Status               string          `json:"status,optional"`
	MetadataConfig       json.RawMessage `json:"metadataConfig,optional"`
	SecretRefConfig      json.RawMessage `json:"secretRefConfig,optional"`
	IntentContractConfig json.RawMessage `json:"intentContractConfig,optional"`
	ImpactScopeConfig    json.RawMessage `json:"impactScopeConfig,optional"`
	Remark               string          `json:"remark,optional"`
}

type CreateChannelIntegrationTemplateResp struct {
	Code    string                               `json:"code"`
	Message string                               `json:"message"`
	Data    CreateChannelIntegrationTemplateData `json:"data"`
}

type CreateChannelIntegrationTemplateData struct {
	Id int64 `json:"id"`
}

type UpdateChannelIntegrationTemplateReq struct {
	Id                   int64           `json:"id"`
	TemplateCode         string          `json:"templateCode"`
	TemplateName         string          `json:"templateName"`
	TemplateType         string          `json:"templateType"`
	TargetCode           string          `json:"targetCode"`
	ScopeType            string          `json:"scopeType"`
	PlatformId           int64           `json:"platformId,optional"`
	TenantId             int64           `json:"tenantId,optional"`
	MerchantId           int64           `json:"merchantId,optional"`
	Status               string          `json:"status,optional"`
	MetadataConfig       json.RawMessage `json:"metadataConfig,optional"`
	SecretRefConfig      json.RawMessage `json:"secretRefConfig,optional"`
	IntentContractConfig json.RawMessage `json:"intentContractConfig,optional"`
	ImpactScopeConfig    json.RawMessage `json:"impactScopeConfig,optional"`
	Remark               string          `json:"remark,optional"`
}

type UpdateChannelIntegrationTemplateStatusReq struct {
	Ids    []int64 `json:"ids"`
	Status string  `json:"status"`
}

type QueryChannelIntegrationTemplateDetailReq struct {
	Id int64 `form:"id"`
}

type QueryChannelIntegrationTemplateListReq struct {
	Current      int64  `form:"current,default=1"`
	PageSize     int64  `form:"pageSize,default=10"`
	TemplateName string `form:"templateName,optional"`
	TemplateType string `form:"templateType,optional"`
	TargetCode   string `form:"targetCode,optional"`
	Status       string `form:"status,optional"`
	TenantId     int64  `form:"tenantId,optional"`
	MerchantId   int64  `form:"merchantId,optional"`
}

type QueryChannelIntegrationTemplateDetailResp struct {
	Code    string                               `json:"code"`
	Message string                               `json:"message"`
	Data    ChannelIntegrationTemplateDetailData `json:"data"`
}

type QueryChannelIntegrationTemplateListResp struct {
	Code     string                                `json:"code"`
	Message  string                                `json:"message"`
	Current  int64                                 `json:"current,default=1"`
	Data     []*ChannelIntegrationTemplateListData `json:"data"`
	PageSize int64                                 `json:"pageSize,default=10"`
	Success  bool                                  `json:"success"`
	Total    int64                                 `json:"total"`
}

type ChannelIntegrationTemplateListData struct {
	Id                   int64           `json:"id"`
	TemplateCode         string          `json:"templateCode"`
	TemplateName         string          `json:"templateName"`
	TemplateType         string          `json:"templateType"`
	TargetCode           string          `json:"targetCode"`
	ScopeType            string          `json:"scopeType"`
	PlatformId           int64           `json:"platformId"`
	TenantId             int64           `json:"tenantId"`
	MerchantId           int64           `json:"merchantId"`
	Status               string          `json:"status"`
	MetadataConfig       json.RawMessage `json:"metadataConfig"`
	SecretRefConfig      json.RawMessage `json:"secretRefConfig"`
	IntentContractConfig json.RawMessage `json:"intentContractConfig"`
	ImpactScopeConfig    json.RawMessage `json:"impactScopeConfig"`
	Remark               string          `json:"remark"`
	CreateBy             string          `json:"createBy"`
	CreateTime           string          `json:"createTime"`
	UpdateBy             string          `json:"updateBy"`
	UpdateTime           string          `json:"updateTime"`
	StatusReason         string          `json:"statusReason"`
	BindingTenantCount   int64           `json:"bindingTenantCount"`
	BindingMerchantCount int64           `json:"bindingMerchantCount"`
	BindingSampleLabels  []string        `json:"bindingSampleLabels"`
	LastStatusChangeTime string          `json:"lastStatusChangeTime"`
}

type ChannelIntegrationTemplateDetailData struct {
	Id                   int64           `json:"id"`
	TemplateCode         string          `json:"templateCode"`
	TemplateName         string          `json:"templateName"`
	TemplateType         string          `json:"templateType"`
	TargetCode           string          `json:"targetCode"`
	ScopeType            string          `json:"scopeType"`
	PlatformId           int64           `json:"platformId"`
	TenantId             int64           `json:"tenantId"`
	MerchantId           int64           `json:"merchantId"`
	Status               string          `json:"status"`
	MetadataConfig       json.RawMessage `json:"metadataConfig"`
	SecretRefConfig      json.RawMessage `json:"secretRefConfig"`
	IntentContractConfig json.RawMessage `json:"intentContractConfig"`
	ImpactScopeConfig    json.RawMessage `json:"impactScopeConfig"`
	Remark               string          `json:"remark"`
	CreateBy             string          `json:"createBy"`
	CreateTime           string          `json:"createTime"`
	UpdateBy             string          `json:"updateBy"`
	UpdateTime           string          `json:"updateTime"`
	StatusReason         string          `json:"statusReason"`
	BindingTenantCount   int64           `json:"bindingTenantCount"`
	BindingMerchantCount int64           `json:"bindingMerchantCount"`
	BindingSampleLabels  []string        `json:"bindingSampleLabels"`
	LastStatusChangeTime string          `json:"lastStatusChangeTime"`
}
