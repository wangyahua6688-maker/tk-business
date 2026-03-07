package lottery

import (
	"time"

	"tk-business/internal/dao"
	"tk-business/internal/security"
	"tk-business/internal/userclient"

	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

// Service 彩种/开奖业务服务。
// 作用：
// 1. 聚合 DAO 层数据并输出前端需要的结构；
// 2. 承担投票防刷策略；
// 3. 优先通过 user gRPC 获取评论，失败时本地降级。
type Service struct {
	dao           *dao.LotteryDAO
	voteLimiter   *security.VoteLimiter
	commentClient userclient.Client
	// sceneRedis 为开奖现场整页缓存的 Redis 客户端；为空表示禁用缓存。
	sceneRedis *redis.Client
	// sceneCacheTTL 控制 live-scene 聚合缓存存活时间。
	sceneCacheTTL time.Duration
}

// VoteMeta 投票请求上下文（用于限流与去重指纹）。
type VoteMeta struct {
	DeviceID  string
	ClientIP  string
	UserAgent string
}

// NewService 创建业务服务。
func NewService(lotteryDAO *dao.LotteryDAO, client userclient.Client, sceneRedis *redis.Client, sceneCacheTTL time.Duration) *Service {
	if sceneCacheTTL <= 0 {
		sceneCacheTTL = 15 * time.Second
	}
	return &Service{
		dao:           lotteryDAO,
		voteLimiter:   security.NewVoteLimiter(rate.Every(20*time.Second), 3, 20*time.Minute),
		commentClient: client,
		sceneRedis:    sceneRedis,
		sceneCacheTTL: sceneCacheTTL,
	}
}
