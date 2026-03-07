package security

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type voter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// VoteLimiter 投票频控器。
// 默认策略：每设备每分钟最多提交 3 次请求，超过则拒绝。
type VoteLimiter struct {
	mu      sync.Mutex
	store   map[string]*voter
	rate    rate.Limit
	burst   int
	expired time.Duration
}

func NewVoteLimiter(rps rate.Limit, burst int, expired time.Duration) *VoteLimiter {
	// 初始化内存频控容器：key 为设备指纹/用户标识，value 为令牌桶状态。
	l := &VoteLimiter{
		store:   make(map[string]*voter),
		rate:    rps,
		burst:   burst,
		expired: expired,
	}
	// 后台协程定时清理长时间未访问的 key，避免内存持续增长。
	go l.gcLoop()
	return l
}

func (l *VoteLimiter) Allow(key string) bool {
	// 空 key 直接拒绝，防止所有匿名请求命中同一默认桶。
	if key == "" {
		return false
	}
	// 访问共享 map 需要互斥锁保护。
	l.mu.Lock()
	defer l.mu.Unlock()

	if v, ok := l.store[key]; ok {
		// 已有令牌桶：更新时间并尝试消费令牌。
		v.lastSeen = time.Now()
		return v.limiter.Allow()
	}

	// 首次访问：创建令牌桶并立即尝试一次请求。
	limiter := rate.NewLimiter(l.rate, l.burst)
	l.store[key] = &voter{limiter: limiter, lastSeen: time.Now()}
	return limiter.Allow()
}

func (l *VoteLimiter) gcLoop() {
	// 固定周期清理过期投票者状态。
	ticker := time.NewTicker(3 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		cutoff := time.Now().Add(-l.expired)
		l.mu.Lock()
		for key, v := range l.store {
			// 超过过期时间未访问则删除，释放内存。
			if v.lastSeen.Before(cutoff) {
				delete(l.store, key)
			}
		}
		l.mu.Unlock()
	}
}
