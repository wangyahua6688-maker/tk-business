package lottery

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
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

	// 读取缓存 JSON 字符串。
	raw, err := s.sceneRedis.Get(ctx, liveSceneCacheKey(specialLotteryID)).Result()
	if err != nil || raw == "" {
		return nil, false
	}

	// 反序列化为 map，异常则视为缓存失效。
	payload := map[string]interface{}{}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
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

	// 序列化失败则直接放弃写缓存，避免影响主链路。
	raw, err := json.Marshal(payload)
	if err != nil {
		return
	}
	_ = s.sceneRedis.Set(ctx, liveSceneCacheKey(specialLotteryID), string(raw), ttl).Err()
}
