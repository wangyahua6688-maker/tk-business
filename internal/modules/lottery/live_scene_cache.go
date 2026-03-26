package lottery

import (
	"context"
	"fmt"
	redisx "github.com/wangyahua6688-maker/tk-common/utils/redisx/v8"
	"time"
)

// liveSceneCacheKey 构造 scene 缓存键：按彩种 ID 隔离，0 代表默认彩种。
func liveSceneCacheKey(specialLotteryID uint) string {
	// 返回当前处理结果。
	return fmt.Sprintf("tk:live_scene:page:v1:%d", specialLotteryID)
}

// loadLiveSceneCache 从 Redis 读取 scene 聚合缓存。
func (s *Service) loadLiveSceneCache(ctx context.Context, specialLotteryID uint) (map[string]interface{}, bool) {
	// Redis 未配置时直接跳过缓存路径。
	if s.sceneRedis == nil {
		// 返回当前处理结果。
		return nil, false
	}
	// 定义并初始化当前变量。
	payload := map[string]interface{}{}
	// 定义并初始化当前变量。
	hit, err := redisx.GetJSON(ctx, s.sceneRedis, liveSceneCacheKey(specialLotteryID), &payload)
	// 判断条件并进入对应分支逻辑。
	if err != nil || !hit {
		// 返回当前处理结果。
		return nil, false
	}
	// 返回当前处理结果。
	return payload, true
}

// saveLiveSceneCache 写入 scene 聚合缓存。
func (s *Service) saveLiveSceneCache(ctx context.Context, specialLotteryID uint, payload map[string]interface{}) {
	// Redis 未配置时无需写缓存。
	if s.sceneRedis == nil || len(payload) == 0 {
		// 返回当前处理结果。
		return
	}

	// 使用配置 TTL；若异常配置为 0，则回退 15 秒。
	ttl := s.sceneCacheTTL
	// 判断条件并进入对应分支逻辑。
	if ttl <= 0 {
		// 更新当前变量或字段值。
		ttl = 15 * time.Second
	}

	// 写缓存失败不影响主流程。
	_ = redisx.SetJSON(ctx, s.sceneRedis, liveSceneCacheKey(specialLotteryID), payload, ttl)
}
