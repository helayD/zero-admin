package channel_integration_template

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"google.golang.org/grpc/status"
)

func grpcError(err error) error {
	s, _ := status.FromError(err)
	return errorx.NewDefaultError(s.Message())
}

func normalizeRawJSON(raw json.RawMessage) string {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return "{}"
	}
	var buffer bytes.Buffer
	if err := json.Compact(&buffer, trimmed); err != nil {
		return strings.TrimSpace(string(trimmed))
	}
	return buffer.String()
}

func mapChannelIntegrationTemplateDetail(item *sysclient.ChannelIntegrationTemplateData) types.ChannelIntegrationTemplateDetailData {
	if item == nil {
		return types.ChannelIntegrationTemplateDetailData{}
	}
	summary := parseImpactSummary(item.ImpactScopeConfig)
	return types.ChannelIntegrationTemplateDetailData{
		Id:                   item.Id,
		TemplateCode:         item.TemplateCode,
		TemplateName:         item.TemplateName,
		TemplateType:         item.TemplateType,
		TargetCode:           item.TargetCode,
		ScopeType:            item.ScopeType,
		PlatformId:           item.PlatformId,
		TenantId:             item.TenantId,
		MerchantId:           item.MerchantId,
		Status:               item.Status,
		MetadataConfig:       json.RawMessage(item.MetadataConfig),
		SecretRefConfig:      json.RawMessage(item.SecretRefConfig),
		IntentContractConfig: json.RawMessage(item.IntentContractConfig),
		ImpactScopeConfig:    json.RawMessage(item.ImpactScopeConfig),
		Remark:               item.Remark,
		CreateBy:             item.CreateBy,
		CreateTime:           item.CreateTime,
		UpdateBy:             item.UpdateBy,
		UpdateTime:           item.UpdateTime,
		StatusReason:         summary.StatusReason,
		BindingTenantCount:   summary.BindingTenantCount,
		BindingMerchantCount: summary.BindingMerchantCount,
		BindingSampleLabels:  summary.BindingSampleLabels,
		LastStatusChangeTime: summary.LastStatusChangeTime,
	}
}

func mapChannelIntegrationTemplateListItem(item *sysclient.ChannelIntegrationTemplateData) *types.ChannelIntegrationTemplateListData {
	if item == nil {
		return &types.ChannelIntegrationTemplateListData{}
	}
	summary := parseImpactSummary(item.ImpactScopeConfig)
	return &types.ChannelIntegrationTemplateListData{
		Id:                   item.Id,
		TemplateCode:         item.TemplateCode,
		TemplateName:         item.TemplateName,
		TemplateType:         item.TemplateType,
		TargetCode:           item.TargetCode,
		ScopeType:            item.ScopeType,
		PlatformId:           item.PlatformId,
		TenantId:             item.TenantId,
		MerchantId:           item.MerchantId,
		Status:               item.Status,
		MetadataConfig:       json.RawMessage(item.MetadataConfig),
		SecretRefConfig:      json.RawMessage(item.SecretRefConfig),
		IntentContractConfig: json.RawMessage(item.IntentContractConfig),
		ImpactScopeConfig:    json.RawMessage(item.ImpactScopeConfig),
		Remark:               item.Remark,
		CreateBy:             item.CreateBy,
		CreateTime:           item.CreateTime,
		UpdateBy:             item.UpdateBy,
		UpdateTime:           item.UpdateTime,
		StatusReason:         summary.StatusReason,
		BindingTenantCount:   summary.BindingTenantCount,
		BindingMerchantCount: summary.BindingMerchantCount,
		BindingSampleLabels:  summary.BindingSampleLabels,
		LastStatusChangeTime: summary.LastStatusChangeTime,
	}
}
