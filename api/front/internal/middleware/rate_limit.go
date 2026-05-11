package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logc"
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
// H-4: memberID 只从 JWT 解析后的 ctx 取，禁止使用客户端可伪造的 X-Member-Id header；
//
//	若 ctx 内无 memberID（如未登录的 H5 校验接口），仅按 IP 限流。
func DigitalCardRateLimitMiddleware(rds *redis.Redis) func(http.HandlerFunc) http.HandlerFunc {
	limiter := NewRateLimiter(rds)
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			clientIP := getClientIP(r)
			memberID := memberIDFromCtx(ctx)
			key := fmt.Sprintf("%s:%s", clientIP, memberID)

			allowed, count, err := limiter.Allow(ctx, key)
			if err != nil {
				// H-6: Redis 异常时放行但必须告警，避免无声绕过限流
				logc.Errorf(ctx, "限流中间件 Redis 异常，临时放行: %v, key=%s", err, key)
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
// H-4: memberID 改从 ctx 取，禁止使用客户端伪造的 header
func AbnormalDetectionMiddleware(rds *redis.Redis) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			clientIP := getClientIP(r)
			memberID := memberIDFromCtx(ctx)

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
// 规则：同一 IP 在 5 分钟内 claim 请求超过 20 次。
// H-5: 移除原"URL.Query 空 token probing"分支——/claim 是 POST + JSON body，
//
//	r.URL.Query 永远不含 token，该分支为死代码且对正常请求计数失真；
//	ZADD member 使用 RemoteAddr + nano 时间戳，避免 POST 下 RawQuery 为空导致 ZADD 合并。
func isAbnormalClaim(ctx context.Context, rds *redis.Redis, clientIP string, r *http.Request) bool {
	claimKey := fmt.Sprintf("claim:abnormal:%s", clientIP)
	now := time.Now().Unix()
	nano := time.Now().UnixNano()
	windowStart := now - int64(5*time.Minute.Seconds())

	// 记录本次请求；member 用 nano 时间戳 + 源端口确保唯一，避免高频请求被 ZADD 合并
	uniqueMember := fmt.Sprintf("%d:%d:%s", now, nano, r.RemoteAddr)
	_, _ = rds.ZaddCtx(ctx, claimKey, now, uniqueMember)
	_ = rds.ExpireCtx(ctx, claimKey, int(5*time.Minute.Seconds()))
	_, _ = rds.ZremrangebyscoreCtx(ctx, claimKey, 0, windowStart)

	count, err := rds.ZcardCtx(ctx, claimKey)
	if err != nil {
		logc.Errorf(ctx, "异常领取检测 Redis 异常: %v, ip=%s", err, clientIP)
		return false
	}
	if int(count) > 20 {
		return true
	}
	return false
}

// memberIDFromCtx 从已登录 JWT 上下文取 memberID（同 frontcommon.GetMemberId 的解析逻辑）。
// 故意复制在 middleware 包内以避免 middleware 依赖 frontcommon 形成循环导入。
// 未登录 / 解析失败返回空字符串，调用方按"匿名"处理。
func memberIDFromCtx(ctx context.Context) string {
	raw := ctx.Value("memberId")
	if raw == nil {
		return ""
	}
	if jn, ok := raw.(json.Number); ok {
		return jn.String()
	}
	if s, ok := raw.(string); ok {
		return s
	}
	if i, ok := raw.(int64); ok {
		return fmt.Sprintf("%d", i)
	}
	return ""
}
