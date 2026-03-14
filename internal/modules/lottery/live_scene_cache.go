package lottery

import (
	"context"
	"fmt"
	"time"

	redisx "tk-common/utils/redisx/v9"
)

// liveSceneCacheKey 构造 scene 缓存键：按彩种 ID 隔离，0 代表默认彩种。
func liveSceneCacheKey(specialLotteryID uint) string {
	return fmt.Sprintf("tk:live_scene:page:v1:%d", specialLotteryID)
}

// loadLiveSceneCache 从 Redis 读取 scene 聚合缓存。
func (s *Service) loadLiveSceneCache(ctx context.Context, specialLotteryID uint) (map[string]interface{}, bool) {
	// Redis 未配置时直接跳过缓存路径。
	if s.sceneRedis == nil {
		return nil, false
	}
	payload := map[string]interface{}{}
	hit, err := redisx.GetJSON(ctx, s.sceneRedis, liveSceneCacheKey(specialLotteryID), &payload)
	if err != nil || !hit {
		return nil, false
	}
	return payload, true
}

// saveLiveSceneCache 写入 scene 聚合缓存。
func (s *Service) saveLiveSceneCache(ctx context.Context, specialLotteryID uint, payload map[string]interface{}) {
	// Redis 未配置时无需写缓存。
	if s.sceneRedis == nil || len(payload) == 0 {
		return
	}

	// 使用配置 TTL；若异常配置为 0，则回退 15 秒。
	ttl := s.sceneCacheTTL
	if ttl <= 0 {
		ttl = 15 * time.Second
	}

	// 写缓存失败不影响主流程。
	_ = redisx.SetJSON(ctx, s.sceneRedis, liveSceneCacheKey(specialLotteryID), payload, ttl)
}
