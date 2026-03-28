package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	// IdempotencyKeyPrefix Redis 幂等键前缀
	IdempotencyKeyPrefix = "order:idempotent:"
	// PROCESSING 状态 TTL=30s（处理中）
	ProcessingTTL = 30 * time.Second
	// CompletedTTL 最终状态 TTL=24h（完成/失败）
	CompletedTTL = 24 * time.Hour
	// IdempotencyProcessing 处理中状态标记
	IdempotencyProcessing = "PROCESSING"
)

// IdempotencyState 幂等键三状态机
type IdempotencyState string

const (
	StateProcessing  IdempotencyState = "PROCESSING"
	StateCompleted   IdempotencyState = "COMPLETED"
	StateFailed      IdempotencyState = "FAILED"
)

// IdempotencyResult 幂等命中时的返回结果
type IdempotencyResult struct {
	State   IdempotencyState
	OrderId int64  // COMPLETED 时返回订单ID
	ErrCode string // FAILED 时返回错误码
	ErrMsg  string // FAILED 时返回错误信息
}

// CheckAndSetProcessing 检查幂等键状态，若不存在则设置为 PROCESSING 并返回 false
// 若已存在则返回当前状态和 true（命中）
func CheckAndSetProcessing(ctx context.Context, rds *redis.Redis, idempotencyKey string) (IdempotencyResult, bool, error) {
	key := IdempotencyKeyPrefix + idempotencyKey

	// 使用 SETNX 原子设置 PROCESSING 状态
	set, err := rds.Setnx(key, IdempotencyProcessing, int(ProcessingTTL.Seconds()))
	if err != nil {
		return IdempotencyResult{}, false, fmt.Errorf("Redis Setnx 异常: %w", err)
	}

	if !set {
		// 键已存在，读取当前状态
		val, err := rds.Get(key)
		if err != nil && err != redis.ErrNotFound {
			return IdempotencyResult{}, false, fmt.Errorf("Redis Get 异常: %w", err)
		}
		if err == redis.ErrNotFound {
			// 键已过期，视为新请求，重试设置
			return CheckAndSetProcessing(ctx, rds, idempotencyKey)
		}
		// 解析状态
		return parseState(val, idempotencyKey, rds, key)
	}

	// 设置成功，当前请求继续处理
	return IdempotencyResult{State: StateProcessing}, false, nil
}

// parseState 解析并返回幂等状态信息
func parseState(val string, _ string, _ *redis.Redis, _ string) (IdempotencyResult, bool, error) {
	if val == IdempotencyProcessing {
		// 另一个请求正在处理中，30s 后超时视为失败
		return IdempotencyResult{State: StateProcessing}, true, nil
	}
	if len(val) > 0 && val[0] == '{' {
		// COMPLETED 状态，val 格式为 {"orderId":123}
		// 尝试解析 JSON
		var result struct {
			OrderId int64 `json:"orderId"`
		}
		if err := json.Unmarshal([]byte(val), &result); err == nil && result.OrderId > 0 {
			return IdempotencyResult{State: StateCompleted, OrderId: result.OrderId}, true, nil
		}
		// 兼容：直接是 orderId 数字字符串
		var orderId int64
		if _, scanErr := fmt.Sscanf(val, "%d", &orderId); scanErr == nil && orderId > 0 {
			return IdempotencyResult{State: StateCompleted, OrderId: orderId}, true, nil
		}
	}
	if len(val) >= 7 && val[:7] == "FAILED:" {
		// FAILED 状态，val 格式为 "FAILED:{errCode}:{errMsg}"
		rest := val[7:]
		colonIdx := -1
		for i, c := range rest {
			if c == ':' {
				colonIdx = i
				break
			}
		}
		if colonIdx >= 0 {
			return IdempotencyResult{
				State:   StateFailed,
				ErrCode: rest[:colonIdx],
				ErrMsg:  rest[colonIdx+1:],
			}, true, nil
		}
		return IdempotencyResult{State: StateFailed, ErrCode: rest, ErrMsg: ""}, true, nil
	}
	// 未知状态，视为新请求
	return IdempotencyResult{State: StateProcessing}, false, nil
}

// MarkCompleted 将幂等键标记为完成状态，存储订单ID
func MarkCompleted(ctx context.Context, rds *redis.Redis, idempotencyKey string, orderId int64) error {
	key := IdempotencyKeyPrefix + idempotencyKey
	val := fmt.Sprintf("{\"orderId\":%d}", orderId)
	return rds.SetExpire(key, val, int(CompletedTTL.Seconds()))
}

// MarkFailed 将幂等键标记为失败状态，存储错误码和错误信息
func MarkFailed(ctx context.Context, rds *redis.Redis, idempotencyKey string, errCode, errMsg string) error {
	key := IdempotencyKeyPrefix + idempotencyKey
	val := fmt.Sprintf("FAILED:%s:%s", errCode, errMsg)
	return rds.SetExpire(key, val, int(CompletedTTL.Seconds()))
}
