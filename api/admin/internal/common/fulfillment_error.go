package common

import (
	"strings"

	"google.golang.org/grpc/status"
)

// FulfillmentValidationError 履约模式校验错误（前端展示用）
type FulfillmentValidationError struct {
	Title       string                          `json:"title"`
	Description string                          `json:"description"`
	Errors      []*FulfillmentErrorDetail       `json:"errors,omitempty"`
}

// FulfillmentErrorDetail 校验错误详情
type FulfillmentErrorDetail struct {
	Code       string `json:"code"`
	Field      string `json:"field"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion"`
}

// ParseFulfillmentValidationErrors 解析履约模式校验错误
// 将 gRPC 错误转换为结构化错误
func ParseFulfillmentValidationErrors(err error) *FulfillmentValidationError {
	if err == nil {
		return nil
	}

	s, ok := status.FromError(err)
	if !ok {
		return &FulfillmentValidationError{
			Title:       "商品保存失败",
			Description: err.Error(),
		}
	}

	message := s.Message()

	// 检查是否是结构化校验错误
	if strings.Contains(message, "FULFILLMENT_") {
		return parseStructuredError(message)
	}

	// 普通错误
	return &FulfillmentValidationError{
		Title:       "商品保存失败",
		Description: message,
	}
}

// parseStructuredError 解析结构化错误消息
func parseStructuredError(message string) *FulfillmentValidationError {
	result := &FulfillmentValidationError{
		Title:  "履约模式配置校验失败",
		Errors: make([]*FulfillmentErrorDetail, 0),
	}

	// 解析普通格式的错误（如 "[FULFILLMENT_RULE_DISABLED] 发卡规则[123]已禁用 (字段: fulfillment_rule_id, 建议: 请启用该发卡规则)"）
	// 提取错误信息
	if idx := strings.Index(message, "]"); idx > 0 {
		code := message[1:idx]
		rest := message[idx+1:]
		
		detail := &FulfillmentErrorDetail{
			Code: code,
		}
		
		// 提取字段
		if fieldIdx := strings.Index(rest, "字段:"); fieldIdx > 0 {
			fieldEnd := strings.Index(rest[fieldIdx:], ")")
			if fieldEnd > 0 {
				detail.Field = strings.TrimSpace(rest[fieldIdx+6 : fieldIdx+fieldEnd])
			}
		}
		
		// 提取建议
		if suggestIdx := strings.Index(rest, "建议:"); suggestIdx > 0 {
			suggestEnd := strings.Index(rest[suggestIdx:], ")")
			if suggestEnd > 0 {
				detail.Suggestion = strings.TrimSpace(rest[suggestIdx+6 : suggestIdx+suggestEnd])
			}
		}
		
		// 提取消息
		if msgIdx := strings.Index(rest, "("); msgIdx > 0 {
			detail.Message = strings.TrimSpace(rest[:msgIdx])
		} else {
			detail.Message = strings.TrimSpace(rest)
		}
		
		result.Errors = append(result.Errors, detail)
		result.Description = detail.Message
	} else {
		result.Description = message
	}

	return result
}
