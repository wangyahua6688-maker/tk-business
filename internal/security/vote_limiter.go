package security

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// 定义当前类型结构。
type voter struct {
	// 处理当前语句逻辑。
	limiter *rate.Limiter
	// 处理当前语句逻辑。
	lastSeen time.Time
}

// VoteLimiter 投票频控器。
// 默认策略：每设备每分钟最多提交 3 次请求，超过则拒绝。
type VoteLimiter struct {
	// 处理当前语句逻辑。
	mu sync.Mutex
	// 处理当前语句逻辑。
	store map[string]*voter
	// 处理当前语句逻辑。
	rate rate.Limit
	// 处理当前语句逻辑。
	burst int
	// 处理当前语句逻辑。
	expired time.Duration
}

// NewVoteLimiter 创建VoteLimiter实例。
func NewVoteLimiter(rps rate.Limit, burst int, expired time.Duration) *VoteLimiter {
	// 初始化内存频控容器：key 为设备指纹/用户标识，value 为令牌桶状态。
	l := &VoteLimiter{
		// 调用make完成当前处理。
		store: make(map[string]*voter),
		// 处理当前语句逻辑。
		rate: rps,
		// 处理当前语句逻辑。
		burst: burst,
		// 处理当前语句逻辑。
		expired: expired,
	}
	// 后台协程定时清理长时间未访问的 key，避免内存持续增长。
	go l.gcLoop()
	// 返回当前处理结果。
	return l
}

// Allow 处理Allow相关逻辑。
func (l *VoteLimiter) Allow(key string) bool {
	// 空 key 直接拒绝，防止所有匿名请求命中同一默认桶。
	if key == "" {
		// 返回当前处理结果。
		return false
	}
	// 访问共享 map 需要互斥锁保护。
	l.mu.Lock()
	// 注册延迟执行逻辑。
	defer l.mu.Unlock()

	// 判断条件并进入对应分支逻辑。
	if v, ok := l.store[key]; ok {
		// 已有令牌桶：更新时间并尝试消费令牌。
		v.lastSeen = time.Now()
		// 返回当前处理结果。
		return v.limiter.Allow()
	}

	// 首次访问：创建令牌桶并立即尝试一次请求。
	limiter := rate.NewLimiter(l.rate, l.burst)
	// 更新当前变量或字段值。
	l.store[key] = &voter{limiter: limiter, lastSeen: time.Now()}
	// 返回当前处理结果。
	return limiter.Allow()
}

// gcLoop 处理gcLoop相关逻辑。
func (l *VoteLimiter) gcLoop() {
	// 固定周期清理过期投票者状态。
	ticker := time.NewTicker(3 * time.Minute)
	// 注册延迟执行逻辑。
	defer ticker.Stop()
	// 循环处理当前数据集合。
	for range ticker.C {
		// 定义并初始化当前变量。
		cutoff := time.Now().Add(-l.expired)
		// 调用l.mu.Lock完成当前处理。
		l.mu.Lock()
		// 循环处理当前数据集合。
		for key, v := range l.store {
			// 超过过期时间未访问则删除，释放内存。
			if v.lastSeen.Before(cutoff) {
				// 调用delete完成当前处理。
				delete(l.store, key)
			}
		}
		// 调用l.mu.Unlock完成当前处理。
		l.mu.Unlock()
	}
}
