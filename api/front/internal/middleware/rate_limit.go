package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	// RateLimitPrefix Redis 限流键前缀
	RateLimitPrefix = "rate_limit:digital_card:"
	// DefaultRateLimit 默认限流次数（每分钟）
	DefaultRateLimit = 10
	// RateLimitWindow 限流窗口（分钟）
	RateLimitWindow = 1 * time.Minute
	// DefaultBurst 默认突发容量
	DefaultBurst = 5
)

// RateLimiter 限流器配置
type RateLimiter struct {
	Redis     *redis.Redis
	KeyPrefix string
	Limit     int
	Window    time.Duration
	Burst     int
}

// NewRateLimiter 创建限流器
func NewRateLimiter(rds *redis.Redis) *RateLimiter {
	return &RateLimiter{
		Redis:     rds,
		KeyPrefix: RateLimitPrefix,
		Limit:     DefaultRateLimit,
		Window:    RateLimitWindow,
		Burst:     DefaultBurst,
	}
}

// Allow 检查是否允许请求通过
// 使用滑动窗口计数器算法
func (rl *RateLimiter) Allow(ctx context.Context, key string) (bool, int, error) {
	fullKey := rl.KeyPrefix + key
	now := time.Now().Unix()
	windowStart := now - int64(rl.Window.Seconds())

	// 移除窗口外的旧计数
	_, err := rl.Redis.ZremrangebyscoreCtx(ctx, fullKey, 0, windowStart)
	if err != nil {
		return false, 0, fmt.Errorf("Redis ZremrangeByScore 异常: %w", err)
	}

	// 添加当前时间戳
	_, err = rl.Redis.ZaddCtx(ctx, fullKey, now, fmt.Sprintf("%d:%d", now, now))
	if err != nil {
		return false, 0, fmt.Errorf("Redis Zadd 异常: %w", err)
	}

	// 获取当前窗口内的计数
	count, err := rl.Redis.ZcardCtx(ctx, fullKey)
	if err != nil {
		return false, 0, fmt.Errorf("Redis Zcard 异常: %w", err)
	}

	// 设置键过期时间
	err = rl.Redis.ExpireCtx(ctx, fullKey, int(rl.Window.Seconds()))
	if err != nil {
		return false, 0, fmt.Errorf("Redis Expire 异常: %w", err)
	}

	// 检查是否超过限流阈值
	if int(count) > rl.Limit+rl.Burst {
		return false, int(count), nil
	}

	return true, int(count), nil
}

// AllowWithBurst 检查是否允许请求通过（带突发容量）
func (rl *RateLimiter) AllowWithBurst(ctx context.Context, key string, burst int) (bool, int, error) {
	fullKey := rl.KeyPrefix + key
	now := time.Now().Unix()
	windowStart := now - int64(rl.Window.Seconds())

	// 移除窗口外的旧计数
	_, err := rl.Redis.ZremrangebyscoreCtx(ctx, fullKey, 0, windowStart)
	if err != nil {
		return false, 0, fmt.Errorf("Redis ZremrangeByScore 异常: %w", err)
	}

	// 添加当前时间戳
	_, err = rl.Redis.ZaddCtx(ctx, fullKey, now, fmt.Sprintf("%d:%d", now, now))
	if err != nil {
		return false, 0, fmt.Errorf("Redis Zadd 异常: %w", err)
	}

	// 获取当前窗口内的计数
	count, err := rl.Redis.ZcardCtx(ctx, fullKey)
	if err != nil {
		return false, 0, fmt.Errorf("Redis Zcard 异常: %w", err)
	}

	// 设置键过期时间
	err = rl.Redis.ExpireCtx(ctx, fullKey, int(rl.Window.Seconds()))
	if err != nil {
		return false, 0, fmt.Errorf("Redis Expire 异常: %w", err)
	}

	if int(count) > rl.Limit+burst {
		return false, int(count), nil
	}

	return true, int(count), nil
}

// DigitalCardRateLimitMiddleware 提货卡接口限流中间件
func DigitalCardRateLimitMiddleware(rds *redis.Redis) func(http.HandlerFunc) http.HandlerFunc {
	limiter := NewRateLimiter(rds)
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			// 使用 IP + 用户ID 作为限流键
			clientIP := getClientIP(r)
			memberID := r.Header.Get("X-Member-Id")
			key := fmt.Sprintf("%s:%s", clientIP, memberID)

			allowed, count, err := limiter.Allow(ctx, key)
			if err != nil {
				// Redis 异常时放行，避免影响正常业务
				next(w, r)
				return
			}

			if !allowed {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limiter.Limit))
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("X-RateLimit-Used", fmt.Sprintf("%d", count))
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"code":429,"message":"请求过于频繁，请稍后再试"}`))
				return
			}

			// 设置限流响应头
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limiter.Limit))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", limiter.Limit+limiter.Burst-count))
			w.Header().Set("X-RateLimit-Used", fmt.Sprintf("%d", count))

			next(w, r)
		}
	}
}

// AbnormalDetectionMiddleware 异常行为检测中间件
// 检测短时间内大量请求、异常IP等
func AbnormalDetectionMiddleware(rds *redis.Redis) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			clientIP := getClientIP(r)
			memberID := r.Header.Get("X-Member-Id")

			// 检查IP是否被封禁
			if isBlocked(ctx, rds, clientIP) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"code":403,"message":"您的IP已被临时封禁，请稍后再试"}`))
				return
			}

			// 检查用户是否被封禁
			if memberID != "" && isUserBlocked(ctx, rds, memberID) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"code":403,"message":"您的账号已被临时封禁，请联系客服"}`))
				return
			}

			// 对 /claim 接口进行异常行为检测
			if strings.Contains(r.URL.Path, "/claim") {
				if isAbnormalClaim(ctx, rds, clientIP, r) {
					_ = BlockIP(ctx, rds, clientIP, 1*time.Hour)
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusForbidden)
					w.Write([]byte(`{"code":403,"message":"检测到异常领取行为，您的IP已被临时封禁"}`))
					return
				}
			}

			next(w, r)
		}
	}
}

// isBlocked 检查IP是否被封禁
func isBlocked(ctx context.Context, rds *redis.Redis, ip string) bool {
	key := fmt.Sprintf("block:ip:%s", ip)
	val, err := rds.GetCtx(ctx, key)
	if err != nil {
		return false
	}
	return val != ""
}

// isUserBlocked 检查用户是否被封禁
func isUserBlocked(ctx context.Context, rds *redis.Redis, memberID string) bool {
	key := fmt.Sprintf("block:user:%s", memberID)
	val, err := rds.GetCtx(ctx, key)
	if err != nil {
		return false
	}
	return val != ""
}

// BlockIP 封禁IP
func BlockIP(ctx context.Context, rds *redis.Redis, ip string, duration time.Duration) error {
	key := fmt.Sprintf("block:ip:%s", ip)
	return rds.SetexCtx(ctx, key, "1", int(duration.Seconds()))
}

// BlockUser 封禁用户
func BlockUser(ctx context.Context, rds *redis.Redis, memberID string, duration time.Duration) error {
	key := fmt.Sprintf("block:user:%s", memberID)
	return rds.SetexCtx(ctx, key, "1", int(duration.Seconds()))
}

// getClientIP 获取真实客户端 IP，优先读取 X-Forwarded-For / X-Real-IP，兼容反向代理
func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}
	if xri := r.Header.Get("X-Real-Ip"); xri != "" {
		return xri
	}
	return r.RemoteAddr
}

// isAbnormalClaim 检测异常领取行为
// 规则：同一 IP 在 5 分钟内 claim 请求超过 20 次，或连续使用空 token  probing
func isAbnormalClaim(ctx context.Context, rds *redis.Redis, clientIP string, r *http.Request) bool {
	claimKey := fmt.Sprintf("claim:abnormal:%s", clientIP)
	now := time.Now().Unix()
	windowStart := now - int64(5*time.Minute.Seconds())

	// 记录本次请求
	_, _ = rds.ZaddCtx(ctx, claimKey, now, fmt.Sprintf("%d:%s", now, r.URL.RawQuery))
	_ = rds.ExpireCtx(ctx, claimKey, int(5*time.Minute.Seconds()))
	_, _ = rds.ZremrangebyscoreCtx(ctx, claimKey, 0, windowStart)

	count, err := rds.ZcardCtx(ctx, claimKey)
	if err != nil {
		return false
	}
	if int(count) > 20 {
		return true
	}

	// 检测连续空 token  probing（5 分钟内超过 5 次空 token）
	token := r.URL.Query().Get("token")
	if token == "" {
		emptyKey := fmt.Sprintf("claim:empty:%s", clientIP)
		emptyCount, _ := rds.IncrCtx(ctx, emptyKey)
		_ = rds.ExpireCtx(ctx, emptyKey, int(5*time.Minute.Seconds()))
		if emptyCount > 5 {
			return true
		}
	}

	return false
}
