package sysclient

import "fmt"

type ChannelIntegrationTemplateData struct {
	Id                   int64  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	TemplateCode         string `protobuf:"bytes,2,opt,name=template_code,json=templateCode,proto3" json:"template_code,omitempty"`
	TemplateName         string `protobuf:"bytes,3,opt,name=template_name,json=templateName,proto3" json:"template_name,omitempty"`
	TemplateType         string `protobuf:"bytes,4,opt,name=template_type,json=templateType,proto3" json:"template_type,omitempty"`
	TargetCode           string `protobuf:"bytes,5,opt,name=target_code,json=targetCode,proto3" json:"target_code,omitempty"`
	ScopeType            string `protobuf:"bytes,6,opt,name=scope_type,json=scopeType,proto3" json:"scope_type,omitempty"`
	PlatformId           int64  `protobuf:"varint,7,opt,name=platform_id,json=platformId,proto3" json:"platform_id,omitempty"`
	TenantId             int64  `protobuf:"varint,8,opt,name=tenant_id,json=tenantId,proto3" json:"tenant_id,omitempty"`
	MerchantId           int64  `protobuf:"varint,9,opt,name=merchant_id,json=merchantId,proto3" json:"merchant_id,omitempty"`
	Status               string `protobuf:"bytes,10,opt,name=status,proto3" json:"status,omitempty"`
	MetadataConfig       string `protobuf:"bytes,11,opt,name=metadata_config,json=metadataConfig,proto3" json:"metadata_config,omitempty"`
	SecretRefConfig      string `protobuf:"bytes,12,opt,name=secret_ref_config,json=secretRefConfig,proto3" json:"secret_ref_config,omitempty"`
	IntentContractConfig string `protobuf:"bytes,13,opt,name=intent_contract_config,json=intentContractConfig,proto3" json:"intent_contract_config,omitempty"`
	ImpactScopeConfig    string `protobuf:"bytes,14,opt,name=impact_scope_config,json=impactScopeConfig,proto3" json:"impact_scope_config,omitempty"`
	Remark               string `protobuf:"bytes,15,opt,name=remark,proto3" json:"remark,omitempty"`
	CreateBy             string `protobuf:"bytes,16,opt,name=create_by,json=createBy,proto3" json:"create_by,omitempty"`
	CreateTime           string `protobuf:"bytes,17,opt,name=create_time,json=createTime,proto3" json:"create_time,omitempty"`
	UpdateBy             string `protobuf:"bytes,18,opt,name=update_by,json=updateBy,proto3" json:"update_by,omitempty"`
	UpdateTime           string `protobuf:"bytes,19,opt,name=update_time,json=updateTime,proto3" json:"update_time,omitempty"`
}

func (x *ChannelIntegrationTemplateData) Reset()         { *x = ChannelIntegrationTemplateData{} }
func (x *ChannelIntegrationTemplateData) String() string { return fmt.Sprintf("%+v", *x) }
func (*ChannelIntegrationTemplateData) ProtoMessage()    {}

type CreateChannelIntegrationTemplateReq struct {
	TemplateCode         string `protobuf:"bytes,1,opt,name=template_code,json=templateCode,proto3" json:"template_code,omitempty"`
	TemplateName         string `protobuf:"bytes,2,opt,name=template_name,json=templateName,proto3" json:"template_name,omitempty"`
	TemplateType         string `protobuf:"bytes,3,opt,name=template_type,json=templateType,proto3" json:"template_type,omitempty"`
	TargetCode           string `protobuf:"bytes,4,opt,name=target_code,json=targetCode,proto3" json:"target_code,omitempty"`
	ScopeType            string `protobuf:"bytes,5,opt,name=scope_type,json=scopeType,proto3" json:"scope_type,omitempty"`
	PlatformId           int64  `protobuf:"varint,6,opt,name=platform_id,json=platformId,proto3" json:"platform_id,omitempty"`
	TenantId             int64  `protobuf:"varint,7,opt,name=tenant_id,json=tenantId,proto3" json:"tenant_id,omitempty"`
	MerchantId           int64  `protobuf:"varint,8,opt,name=merchant_id,json=merchantId,proto3" json:"merchant_id,omitempty"`
	Status               string `protobuf:"bytes,9,opt,name=status,proto3" json:"status,omitempty"`
	MetadataConfig       string `protobuf:"bytes,10,opt,name=metadata_config,json=metadataConfig,proto3" json:"metadata_config,omitempty"`
	SecretRefConfig      string `protobuf:"bytes,11,opt,name=secret_ref_config,json=secretRefConfig,proto3" json:"secret_ref_config,omitempty"`
	IntentContractConfig string `protobuf:"bytes,12,opt,name=intent_contract_config,json=intentContractConfig,proto3" json:"intent_contract_config,omitempty"`
	ImpactScopeConfig    string `protobuf:"bytes,13,opt,name=impact_scope_config,json=impactScopeConfig,proto3" json:"impact_scope_config,omitempty"`
	Remark               string `protobuf:"bytes,14,opt,name=remark,proto3" json:"remark,omitempty"`
	CreateBy             string `protobuf:"bytes,15,opt,name=create_by,json=createBy,proto3" json:"create_by,omitempty"`
}

func (x *CreateChannelIntegrationTemplateReq) Reset()         { *x = CreateChannelIntegrationTemplateReq{} }
func (x *CreateChannelIntegrationTemplateReq) String() string { return fmt.Sprintf("%+v", *x) }
func (*CreateChannelIntegrationTemplateReq) ProtoMessage()    {}

type CreateChannelIntegrationTemplateResp struct {
	Id int64 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *CreateChannelIntegrationTemplateResp) Reset()         { *x = CreateChannelIntegrationTemplateResp{} }
func (x *CreateChannelIntegrationTemplateResp) String() string { return fmt.Sprintf("%+v", *x) }
func (*CreateChannelIntegrationTemplateResp) ProtoMessage()    {}

type UpdateChannelIntegrationTemplateReq struct {
	Id                   int64  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	TemplateCode         string `protobuf:"bytes,2,opt,name=template_code,json=templateCode,proto3" json:"template_code,omitempty"`
	TemplateName         string `protobuf:"bytes,3,opt,name=template_name,json=templateName,proto3" json:"template_name,omitempty"`
	TemplateType         string `protobuf:"bytes,4,opt,name=template_type,json=templateType,proto3" json:"template_type,omitempty"`
	TargetCode           string `protobuf:"bytes,5,opt,name=target_code,json=targetCode,proto3" json:"target_code,omitempty"`
	ScopeType            string `protobuf:"bytes,6,opt,name=scope_type,json=scopeType,proto3" json:"scope_type,omitempty"`
	PlatformId           int64  `protobuf:"varint,7,opt,name=platform_id,json=platformId,proto3" json:"platform_id,omitempty"`
	TenantId             int64  `protobuf:"varint,8,opt,name=tenant_id,json=tenantId,proto3" json:"tenant_id,omitempty"`
	MerchantId           int64  `protobuf:"varint,9,opt,name=merchant_id,json=merchantId,proto3" json:"merchant_id,omitempty"`
	Status               string `protobuf:"bytes,10,opt,name=status,proto3" json:"status,omitempty"`
	MetadataConfig       string `protobuf:"bytes,11,opt,name=metadata_config,json=metadataConfig,proto3" json:"metadata_config,omitempty"`
	SecretRefConfig      string `protobuf:"bytes,12,opt,name=secret_ref_config,json=secretRefConfig,proto3" json:"secret_ref_config,omitempty"`
	IntentContractConfig string `protobuf:"bytes,13,opt,name=intent_contract_config,json=intentContractConfig,proto3" json:"intent_contract_config,omitempty"`
	ImpactScopeConfig    string `protobuf:"bytes,14,opt,name=impact_scope_config,json=impactScopeConfig,proto3" json:"impact_scope_config,omitempty"`
	Remark               string `protobuf:"bytes,15,opt,name=remark,proto3" json:"remark,omitempty"`
	UpdateBy             string `protobuf:"bytes,16,opt,name=update_by,json=updateBy,proto3" json:"update_by,omitempty"`
}

func (x *UpdateChannelIntegrationTemplateReq) Reset()         { *x = UpdateChannelIntegrationTemplateReq{} }
func (x *UpdateChannelIntegrationTemplateReq) String() string { return fmt.Sprintf("%+v", *x) }
func (*UpdateChannelIntegrationTemplateReq) ProtoMessage()    {}

type UpdateChannelIntegrationTemplateResp struct {
	Pong string `protobuf:"bytes,1,opt,name=pong,proto3" json:"pong,omitempty"`
}

func (x *UpdateChannelIntegrationTemplateResp) Reset()         { *x = UpdateChannelIntegrationTemplateResp{} }
func (x *UpdateChannelIntegrationTemplateResp) String() string { return fmt.Sprintf("%+v", *x) }
func (*UpdateChannelIntegrationTemplateResp) ProtoMessage()    {}

type UpdateChannelIntegrationTemplateStatusReq struct {
	Ids      []int64 `protobuf:"varint,1,rep,packed,name=ids,proto3" json:"ids,omitempty"`
	Status   string  `protobuf:"bytes,2,opt,name=status,proto3" json:"status,omitempty"`
	UpdateBy string  `protobuf:"bytes,3,opt,name=update_by,json=updateBy,proto3" json:"update_by,omitempty"`
}

func (x *UpdateChannelIntegrationTemplateStatusReq) Reset() {
	*x = UpdateChannelIntegrationTemplateStatusReq{}
}
func (x *UpdateChannelIntegrationTemplateStatusReq) String() string { return fmt.Sprintf("%+v", *x) }
func (*UpdateChannelIntegrationTemplateStatusReq) ProtoMessage()    {}

type UpdateChannelIntegrationTemplateStatusResp struct {
	Pong string `protobuf:"bytes,1,opt,name=pong,proto3" json:"pong,omitempty"`
}

func (x *UpdateChannelIntegrationTemplateStatusResp) Reset() {
	*x = UpdateChannelIntegrationTemplateStatusResp{}
}
func (x *UpdateChannelIntegrationTemplateStatusResp) String() string {
	return fmt.Sprintf("%+v", *x)
}
func (*UpdateChannelIntegrationTemplateStatusResp) ProtoMessage() {}

type QueryChannelIntegrationTemplateDetailReq struct {
	Id int64 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *QueryChannelIntegrationTemplateDetailReq) Reset() {
	*x = QueryChannelIntegrationTemplateDetailReq{}
}
func (x *QueryChannelIntegrationTemplateDetailReq) String() string { return fmt.Sprintf("%+v", *x) }
func (*QueryChannelIntegrationTemplateDetailReq) ProtoMessage()    {}

type QueryChannelIntegrationTemplateDetailResp struct {
	Data *ChannelIntegrationTemplateData `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *QueryChannelIntegrationTemplateDetailResp) Reset() {
	*x = QueryChannelIntegrationTemplateDetailResp{}
}
func (x *QueryChannelIntegrationTemplateDetailResp) String() string { return fmt.Sprintf("%+v", *x) }
func (*QueryChannelIntegrationTemplateDetailResp) ProtoMessage()    {}

type QueryChannelIntegrationTemplateListReq struct {
	PageNum      int64  `protobuf:"varint,1,opt,name=page_num,json=pageNum,proto3" json:"page_num,omitempty"`
	PageSize     int64  `protobuf:"varint,2,opt,name=page_size,json=pageSize,proto3" json:"page_size,omitempty"`
	TemplateName string `protobuf:"bytes,3,opt,name=template_name,json=templateName,proto3" json:"template_name,omitempty"`
	TemplateType string `protobuf:"bytes,4,opt,name=template_type,json=templateType,proto3" json:"template_type,omitempty"`
	TargetCode   string `protobuf:"bytes,5,opt,name=target_code,json=targetCode,proto3" json:"target_code,omitempty"`
	Status       string `protobuf:"bytes,6,opt,name=status,proto3" json:"status,omitempty"`
	ScopeType    string `protobuf:"bytes,7,opt,name=scope_type,json=scopeType,proto3" json:"scope_type,omitempty"`
	PlatformId   int64  `protobuf:"varint,8,opt,name=platform_id,json=platformId,proto3" json:"platform_id,omitempty"`
	TenantId     int64  `protobuf:"varint,9,opt,name=tenant_id,json=tenantId,proto3" json:"tenant_id,omitempty"`
	MerchantId   int64  `protobuf:"varint,10,opt,name=merchant_id,json=merchantId,proto3" json:"merchant_id,omitempty"`
}

func (x *QueryChannelIntegrationTemplateListReq) Reset() {
	*x = QueryChannelIntegrationTemplateListReq{}
}
func (x *QueryChannelIntegrationTemplateListReq) String() string { return fmt.Sprintf("%+v", *x) }
func (*QueryChannelIntegrationTemplateListReq) ProtoMessage()    {}

type QueryChannelIntegrationTemplateListResp struct {
	Total int64                             `protobuf:"varint,1,opt,name=total,proto3" json:"total,omitempty"`
	List  []*ChannelIntegrationTemplateData `protobuf:"bytes,2,rep,name=list,proto3" json:"list,omitempty"`
}

func (x *QueryChannelIntegrationTemplateListResp) Reset() {
	*x = QueryChannelIntegrationTemplateListResp{}
}
func (x *QueryChannelIntegrationTemplateListResp) String() string { return fmt.Sprintf("%+v", *x) }
func (*QueryChannelIntegrationTemplateListResp) ProtoMessage()    {}
