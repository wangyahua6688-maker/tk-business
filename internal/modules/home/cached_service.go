// Package home 提供带 Redis 缓存的首页聚合服务。
// 本文件是对 service.go 的增强，新增 BuildOverviewWithCache 方法，
// 调用方应优先使用带缓存版本，保留原始 BuildOverview 作为内部穿透调用。
package home

import (
	"context"
	"errors"
	"time"
	"tk-business/internal/constants/rediskey"

	"github.com/wangyahua6688-maker/tk-common/utils/logx"
	"github.com/wangyahua6688-maker/tk-common/utils/redisx/cmdx"
	redisx "github.com/wangyahua6688-maker/tk-common/utils/redisx/v8"

	"github.com/go-redis/redis/v8"
)

// CachedService 是带 Redis 缓存能力的首页服务。
// 嵌入原始 Service，仅新增缓存层方法，不修改原有逻辑。
type CachedService struct {
	*Service
	// redis 是缓存客户端（可为 nil，nil 时退化为直接查 DB）
	redis *redis.Client
	// overviewTTL 首页概览缓存 TTL（可通过配置注入）
	overviewTTL time.Duration
}

// NewCachedService 创建带缓存的首页服务。
// redisClient 为 nil 时退化为无缓存模式（直接查 DB），服务仍然可用。
func NewCachedService(base *Service, redisClient *redis.Client, overviewTTL time.Duration) *CachedService {
	if overviewTTL <= 0 {
		overviewTTL = rediskey.HomeOverviewTTL
	}
	return &CachedService{
		Service:     base,
		redis:       redisClient,
		overviewTTL: overviewTTL,
	}
}

// BuildOverviewWithCache 读取首页聚合数据，优先从 Redis 缓存返回。
//
// 缓存策略：
//   - 命中缓存：直接返回，不查 DB（绝大多数请求）
//   - 缓存 miss：使用互斥锁（SETNX）确保只有一个协程重建缓存，
//     其余协程短暂等待后重试（防缓存击穿）
//   - Redis 不可用：降级直接查 DB（fail-open，服务不中断）
func (s *CachedService) BuildOverviewWithCache(ctx context.Context) (map[string]interface{}, error) {
	logger := logx.LoggerFromContext(ctx)
	cacheKey := rediskey.KeyHomeOverview()

	// ── 1. 尝试从缓存读取 ──────────────────────────
	if s.redis != nil {
		var cached map[string]interface{}
		hit, err := redisx.GetJSON(ctx, s.redis, cacheKey, &cached)
		if err != nil {
			// Redis 读取报错：记录警告但不中断，继续穿透到 DB
			logger.Warn("home.BuildOverviewWithCache: cache get err=%v, fallback to DB", err)
		} else if hit && len(cached) > 0 {
			logger.Debug("home.BuildOverviewWithCache: cache hit key=%s", cacheKey)
			return cached, nil
		}
	}

	// ── 2. 缓存 miss：加互斥锁防击穿 ─────────────────
	if s.redis != nil {
		// 尝试获取重建锁（TTL = 10s，防止锁持有者崩溃导致死锁）
		lockKey := "lock:" + cacheKey
		lock, lockErr := cmdx.AcquireLock(ctx, s.redis, lockKey, 10*time.Second)

		if lockErr == nil {
			// 成功获取锁：执行重建，完成后释放
			defer func() {
				if releaseErr := lock.Release(ctx, s.redis); releaseErr != nil &&
					!errors.Is(releaseErr, cmdx.ErrLockExpired) {
					logger.Warn("home.BuildOverviewWithCache: lock release err=%v", releaseErr)
				}
			}()
		} else if errors.Is(lockErr, cmdx.ErrLockNotAcquired) {
			// 其他协程正在重建：稍等后再读一次缓存（double-check）
			logger.Debug("home.BuildOverviewWithCache: lock not acquired, waiting for rebuild")
			time.Sleep(200 * time.Millisecond)

			var cached map[string]interface{}
			if hit, _ := redisx.GetJSON(ctx, s.redis, cacheKey, &cached); hit && len(cached) > 0 {
				return cached, nil
			}
			// 等待后仍未命中缓存（极小概率），降级查 DB
		}
		// lockErr 为 Redis 通信错误时直接穿透（不阻塞服务）
	}

	// ── 3. 查询 DB 重建数据 ────────────────────────
	logger.Debug("home.BuildOverviewWithCache: cache miss, querying DB key=%s", cacheKey)
	data, err := s.BuildOverview()
	if err != nil {
		return nil, err
	}

	// ── 4. 异步写入缓存（写缓存失败不影响本次响应）───────
	if s.redis != nil && len(data) > 0 {
		go func() {
			// 使用独立 context 写缓存，避免请求 context 取消影响写入
			writeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if writeErr := redisx.SetJSON(writeCtx, s.redis, cacheKey, data, s.overviewTTL); writeErr != nil {
				logger.Warn("home.BuildOverviewWithCache: cache set err=%v", writeErr)
			} else {
				logger.Debug("home.BuildOverviewWithCache: cache set ok key=%s ttl=%s", cacheKey, s.overviewTTL)
			}
		}()
	}

	return data, nil
}

// InvalidateOverviewCache 主动清除首页概览缓存（后台修改配置后调用）。
// 调用场景：Banner/广播/外链/弹窗等配置变更后，由 tk-admin 通过 gRPC 通知
// 或直接调用此方法使缓存失效，下次请求时重建。
func (s *CachedService) InvalidateOverviewCache(ctx context.Context) error {
	if s.redis == nil {
		return nil
	}
	logger := logx.LoggerFromContext(ctx)
	_, err := cmdx.Del(ctx, s.redis, rediskey.KeyHomeOverview())
	if err != nil {
		logger.Error("home.InvalidateOverviewCache: err=%v", err)
		return err
	}
	logger.Info("home.InvalidateOverviewCache: cache invalidated")
	return nil
}
