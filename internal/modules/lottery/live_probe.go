package lottery

import (
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type liveProbeCacheItem struct {
	hasData   bool
	expiresAt time.Time
}

var (
	liveProbeCache sync.Map
	// 探测请求必须短超时，避免拖慢开奖页主链路（BFF 默认 2s 超时）。
	liveProbeClient = &http.Client{Timeout: 800 * time.Millisecond}
)

// probeLiveStreamAvailable 探测直播流是否可播放：
// 1. 未配置或状态异常直接隐藏；
// 2. 对 m3u8 检查是否存在有效分片或子流信息；
// 3. 对普通视频流根据 Content-Type 判定；
// 4. 内存缓存短 TTL，避免每次请求都探测外部地址。
func probeLiveStreamAvailable(streamURL string) bool {
	url := strings.TrimSpace(streamURL)
	if url == "" {
		return false
	}

	// 1) 优先读取缓存；命中则直接返回。
	now := time.Now()
	if cached, ok := liveProbeCache.Load(url); ok {
		item := cached.(liveProbeCacheItem)
		if now.Before(item.expiresAt) {
			return item.hasData
		}
		// 2) 缓存过期时立即返回旧值，并异步刷新，避免阻塞接口响应。
		liveProbeCache.Store(url, liveProbeCacheItem{
			hasData:   item.hasData,
			expiresAt: now.Add(3 * time.Second),
		})
		go refreshLiveProbe(url)
		return item.hasData
	}

	// 3) 首次请求无缓存：先写短期占位，再异步探测，主链路立即返回 false。
	liveProbeCache.Store(url, liveProbeCacheItem{
		hasData:   false,
		expiresAt: now.Add(3 * time.Second),
	})
	go refreshLiveProbe(url)
	return false
}

// refreshLiveProbe 执行真实网络探测，并将结果写回缓存。
func refreshLiveProbe(url string) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		cacheLiveProbe(url, false, 12*time.Second)
		return
	}
	// 只取首段数据用于探测，不下载完整流。
	req.Header.Set("Range", "bytes=0-4096")

	resp, err := liveProbeClient.Do(req)
	if err != nil {
		cacheLiveProbe(url, false, 12*time.Second)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		cacheLiveProbe(url, false, 12*time.Second)
		return
	}

	contentType := strings.ToLower(strings.TrimSpace(resp.Header.Get("Content-Type")))
	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	body := strings.ToLower(string(bodyBytes))

	hasData := false
	if strings.Contains(body, "#extm3u") {
		hasData = strings.Contains(body, "#extinf") ||
			strings.Contains(body, "#ext-x-stream-inf") ||
			strings.Contains(body, ".ts") ||
			strings.Contains(body, ".m4s")
	} else if strings.Contains(contentType, "video/") {
		hasData = true
	} else if strings.Contains(contentType, "application/octet-stream") && len(bodyBytes) > 0 {
		hasData = true
	}

	ttl := 10 * time.Second
	if hasData {
		ttl = 20 * time.Second
	}
	cacheLiveProbe(url, hasData, ttl)
}

func cacheLiveProbe(url string, hasData bool, ttl time.Duration) {
	liveProbeCache.Store(url, liveProbeCacheItem{
		hasData:   hasData,
		expiresAt: time.Now().Add(ttl),
	})
}
