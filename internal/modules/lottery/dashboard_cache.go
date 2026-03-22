// Package lottery 提供开奖看板的 Redis 缓存层，解决开奖高峰缓存击穿问题。
package lottery

import (
	"context"
	"errors"
	"time"

	"tk-business/internal/constants/rediskey"
	"tk-common/utils/logx"
	"tk-common/utils/redisx/cmdx"
	redisx "tk-common/utils/redisx/v8"
)

// BuildDashboardWithCache 读取开奖看板数据，优先从 Redis 缓存返回。
//
// 缓存策略（防缓存击穿核心方案）：
//  1. 尝试读取缓存 → 命中直接返回（99%+ 请求走此路径）
//  2. 缓存 miss → 竞争互斥锁（SETNX）
//     - 获锁成功 → 重建缓存，释放锁
//     - 获锁失败 → 等待 200ms 后重读缓存（double-check）
//  3. Redis 完全不可用 → 降级直接查 DB（fail-open）
//
// 此方法应替换原 BuildDashboard 在 RPC 层的调用。
func (s *Service) BuildDashboardWithCache(ctx context.Context, sid uint) (map[string]interface{}, error) {
	logger := logx.LoggerFromContext(ctx)
	cacheKey := rediskey.KeyDashboard(sid)
	lockKey := rediskey.KeyDashboardLock(sid)

	// ── 1. 快路径：读取缓存 ────────────────────────
	if s.sceneRedis != nil {
		var cached map[string]interface{}
		hit, err := redisx.GetJSON(ctx, s.sceneRedis, cacheKey, &cached)
		if err != nil {
			logger.Warn("dashboard.cache.Get: sid=%d err=%v, fallback to DB", sid, err)
		} else if hit && len(cached) > 0 {
			logger.Debug("dashboard.cache.Get: sid=%d hit", sid)
			return cached, nil
		}
	}

	// ── 2. 缓存 miss：互斥锁防击穿 ──────────────────
	if s.sceneRedis != nil {
		lock, lockErr := cmdx.AcquireLock(ctx, s.sceneRedis, lockKey, rediskey.DashboardLockTTL)

		if lockErr == nil {
			// 获锁成功：执行 DB 重建
			defer func() {
				if releaseErr := lock.Release(ctx, s.sceneRedis); releaseErr != nil &&
					!errors.Is(releaseErr, cmdx.ErrLockExpired) {
					logger.Warn("dashboard.lock.Release: sid=%d err=%v", sid, releaseErr)
				}
			}()
		} else if errors.Is(lockErr, cmdx.ErrLockNotAcquired) {
			// 其他实例正在重建：等待后重读缓存
			logger.Debug("dashboard.lock: sid=%d not acquired, waiting", sid)
			time.Sleep(200 * time.Millisecond)

			var cached map[string]interface{}
			if hit, _ := redisx.GetJSON(ctx, s.sceneRedis, cacheKey, &cached); hit && len(cached) > 0 {
				return cached, nil
			}
			// 仍未命中（极小概率），穿透查 DB
		}
	}

	// ── 3. 查询 DB 重建看板数据 ────────────────────
	logger.Debug("dashboard.rebuild: sid=%d from DB", sid)
	data, err := s.BuildDashboard(sid)
	if err != nil {
		return nil, err
	}

	// ── 4. 写入缓存（TTL = 30s）────────────────────
	if s.sceneRedis != nil && len(data) > 0 {
		go func() {
			writeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			ttl := rediskey.DashboardTTL
			if s.sceneCacheTTL > 0 {
				ttl = s.sceneCacheTTL // 允许配置覆盖
			}
			if writeErr := redisx.SetJSON(writeCtx, s.sceneRedis, cacheKey, data, ttl); writeErr != nil {
				logger.Warn("dashboard.cache.Set: sid=%d err=%v", sid, writeErr)
			}
		}()
	}

	return data, nil
}

// InvalidateDashboardCache 开奖写入后主动清除看板缓存，使下次请求立即获取最新数据。
// 调用时机：tk-admin 写入新开奖结果后触发。
func (s *Service) InvalidateDashboardCache(ctx context.Context, sid uint) error {
	if s.sceneRedis == nil {
		return nil
	}
	logger := logx.LoggerFromContext(ctx)
	_, err := cmdx.Del(ctx, s.sceneRedis, rediskey.KeyDashboard(sid))
	if err != nil {
		logger.Error("dashboard.cache.Invalidate: sid=%d err=%v", sid, err)
		return err
	}
	logger.Info("dashboard.cache.Invalidate: sid=%d cache cleared", sid)
	return nil
}

// BuildLotteryCardsWithCache 读取图纸卡片列表，优先从缓存返回。
func (s *Service) BuildLotteryCardsWithCache(ctx context.Context, category string) ([]map[string]interface{}, error) {
	logger := logx.LoggerFromContext(ctx)
	cacheKey := rediskey.KeyLotteryCards(category)

	// 1. 读缓存
	if s.sceneRedis != nil {
		var cached []map[string]interface{}
		hit, err := redisx.GetJSON(ctx, s.sceneRedis, cacheKey, &cached)
		if err != nil {
			logger.Warn("lottery.cards.cache.Get: category=%s err=%v", category, err)
		} else if hit {
			logger.Debug("lottery.cards.cache.Get: category=%s hit count=%d", category, len(cached))
			return cached, nil
		}
	}

	// 2. 查 DB
	items, err := s.ListCards(category)
	if err != nil {
		return nil, err
	}

	// 3. 写缓存
	if s.sceneRedis != nil {
		go func() {
			writeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			_ = redisx.SetJSON(writeCtx, s.sceneRedis, cacheKey, items, rediskey.LotteryCardsTTL)
		}()
	}

	return items, nil
}

// InvalidateLotteryCardsCache 图纸配置变更后主动清除卡片列表缓存。
func (s *Service) InvalidateLotteryCardsCache(ctx context.Context, category string) error {
	if s.sceneRedis == nil {
		return nil
	}
	key := rediskey.KeyLotteryCards(category)
	_, err := cmdx.Del(ctx, s.sceneRedis, key)
	return err
}
