package lottery

import (
	"context"
	"strings"
	"sync"
	"time"

	commonhttpx "tk-common/utils/httpx"
)

// 定义当前类型结构。
type liveProbeCacheItem struct {
	// 处理当前语句逻辑。
	hasData bool
	// 处理当前语句逻辑。
	expiresAt time.Time
}

// 声明当前变量。
var (
	// 处理当前语句逻辑。
	liveProbeCache sync.Map
	// 探测请求必须短超时，避免拖慢开奖页主链路（BFF 默认 2s 超时）。
	liveProbeClient = commonhttpx.NewTimeoutClient(800 * time.Millisecond)
)

// probeLiveStreamAvailable 探测直播流是否可播放：
// 1. 未配置或状态异常直接隐藏；
// 2. 对 m3u8 检查是否存在有效分片或子流信息；
// 3. 对普通视频流根据 Content-Type 判定；
// 4. 内存缓存短 TTL，避免每次请求都探测外部地址。
func probeLiveStreamAvailable(streamURL string) bool {
	// 定义并初始化当前变量。
	url := strings.TrimSpace(streamURL)
	// 判断条件并进入对应分支逻辑。
	if url == "" {
		// 返回当前处理结果。
		return false
	}

	// 1) 优先读取缓存；命中则直接返回。
	now := time.Now()
	// 判断条件并进入对应分支逻辑。
	if cached, ok := liveProbeCache.Load(url); ok {
		// 定义并初始化当前变量。
		item := cached.(liveProbeCacheItem)
		// 判断条件并进入对应分支逻辑。
		if now.Before(item.expiresAt) {
			// 返回当前处理结果。
			return item.hasData
		}
		// 2) 缓存过期时立即返回旧值，并异步刷新，避免阻塞接口响应。
		liveProbeCache.Store(url, liveProbeCacheItem{
			// 处理当前语句逻辑。
			hasData: item.hasData,
			// 调用now.Add完成当前处理。
			expiresAt: now.Add(3 * time.Second),
		})
		// 异步启动后台处理任务。
		go refreshLiveProbe(url)
		// 返回当前处理结果。
		return item.hasData
	}

	// 3) 首次请求无缓存：先写短期占位，再异步探测，主链路立即返回 false。
	liveProbeCache.Store(url, liveProbeCacheItem{
		// 处理当前语句逻辑。
		hasData: false,
		// 调用now.Add完成当前处理。
		expiresAt: now.Add(3 * time.Second),
	})
	// 异步启动后台处理任务。
	go refreshLiveProbe(url)
	// 返回当前处理结果。
	return false
}

// refreshLiveProbe 执行真实网络探测，并将结果写回缓存。
func refreshLiveProbe(url string) {
	// 定义并初始化当前变量。
	statusCode, contentTypeRaw, bodyBytes, err := commonhttpx.GetRange(
		// 调用context.Background完成当前处理。
		context.Background(),
		// 处理当前语句逻辑。
		liveProbeClient,
		// 处理当前语句逻辑。
		url,
		// 更新当前变量或字段值。
		"bytes=0-4096",
		// 处理当前语句逻辑。
		8192,
	)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 调用cacheLiveProbe完成当前处理。
		cacheLiveProbe(url, false, 12*time.Second)
		// 返回当前处理结果。
		return
	}

	// 判断条件并进入对应分支逻辑。
	if statusCode >= 400 {
		// 调用cacheLiveProbe完成当前处理。
		cacheLiveProbe(url, false, 12*time.Second)
		// 返回当前处理结果。
		return
	}

	// 定义并初始化当前变量。
	contentType := strings.ToLower(strings.TrimSpace(contentTypeRaw))
	// 定义并初始化当前变量。
	body := strings.ToLower(string(bodyBytes))

	// 定义并初始化当前变量。
	hasData := false
	// 判断条件并进入对应分支逻辑。
	if strings.Contains(body, "#extm3u") {
		// 更新当前变量或字段值。
		hasData = strings.Contains(body, "#extinf") ||
			// 调用strings.Contains完成当前处理。
			strings.Contains(body, "#ext-x-stream-inf") ||
			// 调用strings.Contains完成当前处理。
			strings.Contains(body, ".ts") ||
			// 调用strings.Contains完成当前处理。
			strings.Contains(body, ".m4s")
		// 调用strings.Contains完成当前处理。
	} else if strings.Contains(contentType, "video/") {
		// 更新当前变量或字段值。
		hasData = true
		// 调用strings.Contains完成当前处理。
	} else if strings.Contains(contentType, "application/octet-stream") && len(bodyBytes) > 0 {
		// 更新当前变量或字段值。
		hasData = true
	}

	// 定义并初始化当前变量。
	ttl := 10 * time.Second
	// 判断条件并进入对应分支逻辑。
	if hasData {
		// 更新当前变量或字段值。
		ttl = 20 * time.Second
	}
	// 调用cacheLiveProbe完成当前处理。
	cacheLiveProbe(url, hasData, ttl)
}

// cacheLiveProbe 处理cacheLiveProbe相关逻辑。
func cacheLiveProbe(url string, hasData bool, ttl time.Duration) {
	// 调用liveProbeCache.Store完成当前处理。
	liveProbeCache.Store(url, liveProbeCacheItem{
		// 处理当前语句逻辑。
		hasData: hasData,
		// 调用time.Now完成当前处理。
		expiresAt: time.Now().Add(ttl),
	})
}
