package draw_activity

import (
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"google.golang.org/grpc/status"
)

func grpcDrawError(err error) error {
	if err == nil {
		return nil
	}
	s, ok := status.FromError(err)
	if ok {
		return errorx.NewDefaultError(s.Message())
	}
	return errorx.NewDefaultError(err.Error())
}

func toSMSHomeEntry(input types.DrawHomeEntryConfig) *smsclient.DrawHomeEntryConfig {
	return &smsclient.DrawHomeEntryConfig{
		ShowOnHome:         input.ShowOnHome,
		HomeEntryTitle:     input.HomeEntryTitle,
		HomeEntrySubtitle:  input.HomeEntrySubtitle,
		HomeEntryImage:     input.HomeEntryImage,
		HomeEntrySort:      input.HomeEntrySort,
		LandingTargetType:  input.LandingTargetType,
		LandingTargetValue: input.LandingTargetValue,
		IsEnabled:          input.IsEnabled,
	}
}

func toSMSTemplates(input []types.DrawCardTemplateData) []*smsclient.DrawCardTemplateData {
	result := make([]*smsclient.DrawCardTemplateData, 0, len(input))
	for i := range input {
		item := input[i]
		result = append(result, &smsclient.DrawCardTemplateData{
			Id:                      item.Id,
			TemplateName:            item.TemplateName,
			TemplateCode:            item.TemplateCode,
			CardFaceImage:           item.CardFaceImage,
			CopyrightOwner:          item.CopyrightOwner,
			CopyrightProofSummary:   item.CopyrightProofSummary,
			Rarity:                  item.Rarity,
			IssueLimit:              item.IssueLimit,
			DisplayCopy:             item.DisplayCopy,
			CirculationLimitSummary: item.CirculationLimitSummary,
			DisplayStatus:           item.DisplayStatus,
			ContentAuditStatus:      item.ContentAuditStatus,
			ProviderCode:            item.ProviderCode,
			CredentialRef:           item.CredentialRef,
			Status:                  item.Status,
			AuditStatus:             item.AuditStatus,
		})
	}
	return result
}

func toSMSPools(input []types.DrawPoolData) []*smsclient.DrawPoolData {
	result := make([]*smsclient.DrawPoolData, 0, len(input))
	for i := range input {
		item := input[i]
		mappings := make([]*smsclient.DrawPoolTemplateData, 0, len(item.Templates))
		for j := range item.Templates {
			mapping := item.Templates[j]
			slotIndex := mapping.SlotIndex
			if slotIndex <= 0 {
				slotIndex = int32(j + 1)
			}
			mappings = append(mappings, &smsclient.DrawPoolTemplateData{
				Id:             mapping.Id,
				TemplateId:     mapping.TemplateId,
				TemplateCode:   mapping.TemplateCode,
				TemplateName:   mapping.TemplateName,
				Rarity:         mapping.Rarity,
				SlotIndex:      slotIndex,
				Probability:    mapping.Probability,
				SaleLimit:      mapping.SaleLimit,
				RemainingLimit: mapping.RemainingLimit,
				ConfigLimit:    mapping.ConfigLimit,
			})
		}
		result = append(result, &smsclient.DrawPoolData{
			Id:              item.Id,
			PoolName:        item.PoolName,
			PoolCode:        item.PoolCode,
			ProbabilityRule: item.ProbabilityRule,
			Sort:            item.Sort,
			Status:          item.Status,
			Templates:       mappings,
		})
	}
	return result
}

func toAPITemplates(input []*smsclient.DrawCardTemplateData) []types.DrawCardTemplateData {
	result := make([]types.DrawCardTemplateData, 0, len(input))
	for _, item := range input {
		if item == nil {
			continue
		}
		result = append(result, types.DrawCardTemplateData{
			Id:                      item.Id,
			TemplateName:            item.TemplateName,
			TemplateCode:            item.TemplateCode,
			CardFaceImage:           item.CardFaceImage,
			CopyrightOwner:          item.CopyrightOwner,
			CopyrightProofSummary:   item.CopyrightProofSummary,
			Rarity:                  item.Rarity,
			IssueLimit:              item.IssueLimit,
			DisplayCopy:             item.DisplayCopy,
			CirculationLimitSummary: item.CirculationLimitSummary,
			DisplayStatus:           item.DisplayStatus,
			ContentAuditStatus:      item.ContentAuditStatus,
			ProviderCode:            item.ProviderCode,
			CredentialRef:           item.CredentialRef,
			Status:                  item.Status,
			AuditStatus:             item.AuditStatus,
		})
	}
	return result
}

func toAPIPools(input []*smsclient.DrawPoolData) []types.DrawPoolData {
	result := make([]types.DrawPoolData, 0, len(input))
	for _, item := range input {
		if item == nil {
			continue
		}
		mappings := make([]types.DrawPoolTemplateData, 0, len(item.Templates))
		for _, mapping := range item.Templates {
			if mapping == nil {
				continue
			}
			mappings = append(mappings, types.DrawPoolTemplateData{
				Id:             mapping.Id,
				TemplateId:     mapping.TemplateId,
				TemplateCode:   mapping.TemplateCode,
				TemplateName:   mapping.TemplateName,
				Rarity:         mapping.Rarity,
				SlotIndex:      mapping.SlotIndex,
				Probability:    mapping.Probability,
				SaleLimit:      mapping.SaleLimit,
				RemainingLimit: mapping.RemainingLimit,
				ConfigLimit:    mapping.ConfigLimit,
			})
		}
		result = append(result, types.DrawPoolData{
			Id:              item.Id,
			PoolName:        item.PoolName,
			PoolCode:        item.PoolCode,
			ProbabilityRule: item.ProbabilityRule,
			Sort:            item.Sort,
			Status:          item.Status,
			Templates:       mappings,
		})
	}
	return result
}

func toAPIReadinessItems(input []*smsclient.DrawReadinessItem) []types.DrawReadinessItem {
	result := make([]types.DrawReadinessItem, 0, len(input))
	for _, item := range input {
		if item == nil {
			continue
		}
		result = append(result, types.DrawReadinessItem{
			Code:     item.Code,
			Field:    item.Field,
			Message:  item.Message,
			Blocking: item.Blocking,
		})
	}
	return result
}

func toAPIAudits(input []*smsclient.DrawActivityAuditRecord) []types.DrawActivityAuditRecord {
	result := make([]types.DrawActivityAuditRecord, 0, len(input))
	for _, item := range input {
		if item == nil {
			continue
		}
		result = append(result, types.DrawActivityAuditRecord{
			Id:             item.Id,
			OperationType:  item.OperationType,
			OperatorId:     item.OperatorId,
			OperatorName:   item.OperatorName,
			ApprovalResult: item.ApprovalResult,
			FailureSummary: item.FailureSummary,
			TraceId:        item.TraceId,
			CreateTime:     item.CreateTime,
		})
	}
	return result
}
