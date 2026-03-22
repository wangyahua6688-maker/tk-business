package lottery

import (
	"time"

	"tk-business/internal/dao"
	"tk-business/internal/security"
	"tk-business/internal/userclient"

	"github.com/go-redis/redis/v8"
	"golang.org/x/time/rate"
)

// Service 彩种/开奖业务服务。
// 作用：
// 1. 聚合 DAO 层数据并输出前端需要的结构；
// 2. 承担投票防刷策略；
// 3. 优先通过 user gRPC 获取评论，失败时本地降级。
type Service struct {
	// 处理当前语句逻辑。
	dao *dao.LotteryDAO
	// 处理当前语句逻辑。
	voteLimiter *security.VoteLimiter
	// 处理当前语句逻辑。
	commentClient userclient.Client
	// sceneRedis 为开奖现场整页缓存的 Redis 客户端；为空表示禁用缓存。
	sceneRedis *redis.Client
	// sceneCacheTTL 控制 live-scene 聚合缓存存活时间。
	sceneCacheTTL time.Duration
}

// VoteMeta 投票请求上下文（用于限流与去重指纹）。
type VoteMeta struct {
	// 处理当前语句逻辑。
	DeviceID string
	// 处理当前语句逻辑。
	ClientIP string
	// 处理当前语句逻辑。
	UserAgent string
}

// NewService 创建业务服务。
func NewService(lotteryDAO *dao.LotteryDAO, client userclient.Client, sceneRedis *redis.Client, sceneCacheTTL time.Duration) *Service {
	// 判断条件并进入对应分支逻辑。
	if sceneCacheTTL <= 0 {
		// 更新当前变量或字段值。
		sceneCacheTTL = 15 * time.Second
	}
	// 返回当前处理结果。
	return &Service{
		// 处理当前语句逻辑。
		dao: lotteryDAO,
		// 调用security.NewVoteLimiter完成当前处理。
		voteLimiter: security.NewVoteLimiter(rate.Every(20*time.Second), 3, 20*time.Minute),
		// 处理当前语句逻辑。
		commentClient: client,
		// 处理当前语句逻辑。
		sceneRedis: sceneRedis,
		// 处理当前语句逻辑。
		sceneCacheTTL: sceneCacheTTL,
	}
}
